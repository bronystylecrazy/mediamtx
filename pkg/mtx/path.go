package mtx

import (
	"context"
	"fmt"
	conf2 "github.com/bluenviron/mediamtx/pkg/conf"
	defs2 "github.com/bluenviron/mediamtx/pkg/defs"
	"github.com/bluenviron/mediamtx/pkg/externalcmd"
	"github.com/bluenviron/mediamtx/pkg/hooks"
	"github.com/bluenviron/mediamtx/pkg/logger"
	"github.com/bluenviron/mediamtx/pkg/recorder"
	"github.com/bluenviron/mediamtx/pkg/staticsources"
	"github.com/bluenviron/mediamtx/pkg/stream"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v5/pkg/description"

	"github.com/bluenviron/mediamtx/pkg/utils"
)

type OptionalPath = conf2.OptionalPath

func emptyTimer() *time.Timer {
	t := time.NewTimer(0)
	<-t.C
	return t
}

type pathParent interface {
	logger.Writer
	pathReady(*PathHandler)
	pathNotReady(*PathHandler)
	closePath(*PathHandler)
	AddReader(req defs2.PathAddReaderReq) (defs2.Path, *stream.Stream, error)

	PathHook
}

type pathOnDemandState int

const (
	pathOnDemandStateInitial pathOnDemandState = iota
	pathOnDemandStateWaitingReady
	pathOnDemandStateReady
	pathOnDemandStateClosing
)

type pathAPIPathsListRes struct {
	data  *defs2.APIPathList
	paths map[string]*PathHandler
}

type pathAPIPathsListReq struct {
	res chan pathAPIPathsListRes
}

type pathAPIPathsGetRes struct {
	path *PathHandler
	data *defs2.APIPath
	err  error
}

type pathAPIPathsGetReq struct {
	name string
	res  chan pathAPIPathsGetRes
}

type PathHandler struct {
	parentCtx         context.Context
	logLevel          conf2.LogLevel
	rtspAddress       string
	readTimeout       conf2.Duration
	writeTimeout      conf2.Duration
	writeQueueSize    int
	rtpMaxPayloadSize int
	conf              *conf2.Path
	name              string
	matches           []string
	wg                *sync.WaitGroup
	externalCmdPool   *externalcmd.Pool
	parent            pathParent

	ctx                            context.Context
	ctxCancel                      func()
	confMutex                      sync.RWMutex
	source                         defs2.Source
	publisherQuery                 string
	stream                         *stream.Stream
	recorder                       *recorder.Recorder
	readyTime                      time.Time
	onUnDemandHook                 func(string)
	onNotReadyHook                 func()
	readers                        map[defs2.Reader]struct{}
	describeRequestsOnHold         []defs2.PathDescribeReq
	readerAddRequestsOnHold        []defs2.PathAddReaderReq
	onDemandStaticSourceState      pathOnDemandState
	onDemandStaticSourceReadyTimer *time.Timer
	onDemandStaticSourceCloseTimer *time.Timer
	onDemandPublisherState         pathOnDemandState
	onDemandPublisherReadyTimer    *time.Timer
	onDemandPublisherCloseTimer    *time.Timer

	// in
	chReloadConf              chan *conf2.Path
	chStaticSourceSetReady    chan defs2.PathSourceStaticSetReadyReq
	chStaticSourceSetNotReady chan defs2.PathSourceStaticSetNotReadyReq
	chDescribe                chan defs2.PathDescribeReq
	chAddPublisher            chan defs2.PathAddPublisherReq
	chRemovePublisher         chan defs2.PathRemovePublisherReq
	chAddReader               chan defs2.PathAddReaderReq
	chRemoveReader            chan defs2.PathRemoveReaderReq
	chAPIPathsGet             chan pathAPIPathsGetReq

	// out
	done chan struct{}
}

func (pa *PathHandler) initialize() {
	ctx, ctxCancel := context.WithCancel(pa.parentCtx)

	pa.ctx = ctx
	pa.ctxCancel = ctxCancel
	pa.readers = make(map[defs2.Reader]struct{})
	pa.onDemandStaticSourceReadyTimer = emptyTimer()
	pa.onDemandStaticSourceCloseTimer = emptyTimer()
	pa.onDemandPublisherReadyTimer = emptyTimer()
	pa.onDemandPublisherCloseTimer = emptyTimer()
	pa.chReloadConf = make(chan *conf2.Path)
	pa.chStaticSourceSetReady = make(chan defs2.PathSourceStaticSetReadyReq)
	pa.chStaticSourceSetNotReady = make(chan defs2.PathSourceStaticSetNotReadyReq)
	pa.chDescribe = make(chan defs2.PathDescribeReq)
	pa.chAddPublisher = make(chan defs2.PathAddPublisherReq)
	pa.chRemovePublisher = make(chan defs2.PathRemovePublisherReq)
	pa.chAddReader = make(chan defs2.PathAddReaderReq)
	pa.chRemoveReader = make(chan defs2.PathRemoveReaderReq)
	pa.chAPIPathsGet = make(chan pathAPIPathsGetReq)
	pa.done = make(chan struct{})

	pa.Log(logger.Debug, "created")

	pa.wg.Add(1)
	go pa.run()
}

func (pa *PathHandler) close() {
	pa.ctxCancel()
}

func (pa *PathHandler) wait() {
	<-pa.done
}

// Log implements logger.Writer.
func (pa *PathHandler) Log(level logger.Level, format string, args ...interface{}) {
	pa.parent.Log(level, "[Path "+pa.name+"] "+format, args...)
}

func (pa *PathHandler) Name() string {
	return pa.name
}

func (pa *PathHandler) isReady() bool {
	return pa.stream != nil
}

func (pa *PathHandler) run() {
	defer close(pa.done)
	defer pa.wg.Done()

	if pa.conf.Source == "redirect" {
		pa.source = &sourceRedirect{}
	} else if pa.conf.HasStaticSource() {
		pa.source = &staticsources.Handler{
			Conf:              pa.conf,
			LogLevel:          pa.logLevel,
			ReadTimeout:       pa.readTimeout,
			WriteTimeout:      pa.writeTimeout,
			WriteQueueSize:    pa.writeQueueSize,
			RTPMaxPayloadSize: pa.rtpMaxPayloadSize,
			Matches:           pa.matches,
			PathManager:       pa.parent,
			Parent:            pa,
		}
		pa.source.(*staticsources.Handler).Initialize()

		if !pa.conf.SourceOnDemand {
			pa.source.(*staticsources.Handler).Start(false, "")
		}
	}

	onUnInitHook := hooks.OnInit(hooks.OnInitParams{
		Logger:          pa,
		ExternalCmdPool: pa.externalCmdPool,
		Conf:            pa.conf,
		ExternalCmdEnv:  pa.ExternalCmdEnv(),
	})

	pa.parent.OnInit(pa.ctx, pa)

	err := pa.runInner()

	// call before destroying context
	pa.parent.closePath(pa)

	pa.ctxCancel()

	pa.onDemandStaticSourceReadyTimer.Stop()
	pa.onDemandStaticSourceCloseTimer.Stop()
	pa.onDemandPublisherReadyTimer.Stop()
	pa.onDemandPublisherCloseTimer.Stop()

	onUnInitHook()
	pa.parent.OnUnInit(pa.ctx, pa)

	for _, req := range pa.describeRequestsOnHold {
		req.Res <- defs2.PathDescribeRes{Err: fmt.Errorf("terminated")}
	}

	for _, req := range pa.readerAddRequestsOnHold {
		req.Res <- defs2.PathAddReaderRes{Err: fmt.Errorf("terminated")}
	}

	if pa.stream != nil {
		pa.setNotReady()
	}

	if pa.source != nil {
		if source, ok := pa.source.(*staticsources.Handler); ok {
			if !pa.conf.SourceOnDemand || pa.onDemandStaticSourceState != pathOnDemandStateInitial {
				source.Close("PathHandler is closing")
			}
		} else if source, ok2 := pa.source.(defs2.Publisher); ok2 {
			source.Close()
		}
	}

	if pa.onUnDemandHook != nil {
		pa.onUnDemandHook("PathHandler destroyed")
	}

	pa.Log(logger.Debug, "destroyed: %v", err)
}

func (pa *PathHandler) runInner() error {
	for {
		select {
		case <-pa.onDemandStaticSourceReadyTimer.C:
			pa.doOnDemandStaticSourceReadyTimer()

			if pa.shouldClose() {
				return fmt.Errorf("not in use")
			}

		case <-pa.onDemandStaticSourceCloseTimer.C:
			pa.doOnDemandStaticSourceCloseTimer()

			if pa.shouldClose() {
				return fmt.Errorf("not in use")
			}

		case <-pa.onDemandPublisherReadyTimer.C:
			pa.doOnDemandPublisherReadyTimer()

			if pa.shouldClose() {
				return fmt.Errorf("not in use")
			}

		case <-pa.onDemandPublisherCloseTimer.C:
			pa.doOnDemandPublisherCloseTimer()

		case newConf := <-pa.chReloadConf:
			pa.doReloadConf(newConf)

		case req := <-pa.chStaticSourceSetReady:
			pa.doSourceStaticSetReady(req)

		case req := <-pa.chStaticSourceSetNotReady:
			pa.doSourceStaticSetNotReady(req)

			if pa.shouldClose() {
				return fmt.Errorf("not in use")
			}

		case req := <-pa.chDescribe:
			pa.doDescribe(req)

			if pa.shouldClose() {
				return fmt.Errorf("not in use")
			}

		case req := <-pa.chAddPublisher:
			pa.doAddPublisher(req)

		case req := <-pa.chRemovePublisher:
			pa.doRemovePublisher(req)

			if pa.shouldClose() {
				return fmt.Errorf("not in use")
			}

		case req := <-pa.chAddReader:
			pa.doAddReader(req)

			if pa.shouldClose() {
				return fmt.Errorf("not in use")
			}

		case req := <-pa.chRemoveReader:
			pa.doRemoveReader(req)

		case req := <-pa.chAPIPathsGet:
			pa.doAPIPathsGet(req)

		case <-pa.ctx.Done():
			return fmt.Errorf("terminated")
		}
	}
}

func (pa *PathHandler) doOnDemandStaticSourceReadyTimer() {
	for _, req := range pa.describeRequestsOnHold {
		req.Res <- defs2.PathDescribeRes{Err: fmt.Errorf("source of PathHandler '%s' has timed out", pa.name)}
	}
	pa.describeRequestsOnHold = nil

	for _, req := range pa.readerAddRequestsOnHold {
		req.Res <- defs2.PathAddReaderRes{Err: fmt.Errorf("source of PathHandler '%s' has timed out", pa.name)}
	}
	pa.readerAddRequestsOnHold = nil

	pa.onDemandStaticSourceStop("timed out")
}

func (pa *PathHandler) doOnDemandStaticSourceCloseTimer() {
	pa.setNotReady()
	pa.onDemandStaticSourceStop("not needed by anyone")
}

func (pa *PathHandler) doOnDemandPublisherReadyTimer() {
	for _, req := range pa.describeRequestsOnHold {
		req.Res <- defs2.PathDescribeRes{Err: fmt.Errorf("source of PathHandler '%s' has timed out", pa.name)}
	}
	pa.describeRequestsOnHold = nil

	for _, req := range pa.readerAddRequestsOnHold {
		req.Res <- defs2.PathAddReaderRes{Err: fmt.Errorf("source of PathHandler '%s' has timed out", pa.name)}
	}
	pa.readerAddRequestsOnHold = nil

	pa.onDemandPublisherStop("timed out")
}

func (pa *PathHandler) doOnDemandPublisherCloseTimer() {
	pa.onDemandPublisherStop("not needed by anyone")
}

func (pa *PathHandler) doReloadConf(newConf *conf2.Path) {
	pa.confMutex.Lock()
	oldConf := pa.conf
	pa.conf = newConf
	pa.confMutex.Unlock()

	if pa.conf.HasStaticSource() {
		pa.source.(*staticsources.Handler).ReloadConf(newConf)
	}

	if pa.recorder != nil &&
		(newConf.Record != oldConf.Record ||
			newConf.RecordPath != oldConf.RecordPath ||
			newConf.RecordFormat != oldConf.RecordFormat ||
			newConf.RecordPartDuration != oldConf.RecordPartDuration ||
			newConf.RecordMaxPartSize != oldConf.RecordMaxPartSize ||
			newConf.RecordSegmentDuration != oldConf.RecordSegmentDuration ||
			newConf.RecordDeleteAfter != oldConf.RecordDeleteAfter) {
		pa.recorder.Close()
		pa.recorder = nil
	}

	if newConf.Record && pa.stream != nil && pa.recorder == nil {
		pa.startRecording()
	}
}

func (pa *PathHandler) doSourceStaticSetReady(req defs2.PathSourceStaticSetReadyReq) {
	err := pa.setReady(req.Desc, req.GenerateRTPPackets)
	if err != nil {
		req.Res <- defs2.PathSourceStaticSetReadyRes{Err: err}
		return
	}

	if pa.conf.HasOnDemandStaticSource() {
		pa.onDemandStaticSourceReadyTimer.Stop()
		pa.onDemandStaticSourceReadyTimer = emptyTimer()
		pa.onDemandStaticSourceScheduleClose()
	}

	pa.consumeOnHoldRequests()

	req.Res <- defs2.PathSourceStaticSetReadyRes{Stream: pa.stream}
}

func (pa *PathHandler) doSourceStaticSetNotReady(req defs2.PathSourceStaticSetNotReadyReq) {
	pa.setNotReady()

	// send response before calling onDemandStaticSourceStop()
	// in order to avoid a deadlock due to staticsources.Handler.stop()
	close(req.Res)

	if pa.conf.HasOnDemandStaticSource() && pa.onDemandStaticSourceState != pathOnDemandStateInitial {
		pa.onDemandStaticSourceStop("an error occurred")
	}
}

func (pa *PathHandler) doDescribe(req defs2.PathDescribeReq) {
	if _, ok := pa.source.(*sourceRedirect); ok {
		req.Res <- defs2.PathDescribeRes{
			Redirect: pa.conf.SourceRedirect,
		}
		return
	}

	if pa.stream != nil {
		req.Res <- defs2.PathDescribeRes{
			Stream: pa.stream,
		}
		return
	}

	if pa.conf.HasOnDemandStaticSource() {
		if pa.onDemandStaticSourceState == pathOnDemandStateInitial {
			pa.onDemandStaticSourceStart(req.AccessRequest.Query)
		}
		pa.describeRequestsOnHold = append(pa.describeRequestsOnHold, req)
		return
	}

	if pa.conf.HasOnDemandPublisher() {
		if pa.onDemandPublisherState == pathOnDemandStateInitial {
			pa.onDemandPublisherStart(req.AccessRequest.Query)
		}
		pa.describeRequestsOnHold = append(pa.describeRequestsOnHold, req)
		return
	}

	if pa.conf.Fallback != "" {
		req.Res <- defs2.PathDescribeRes{Redirect: pa.conf.Fallback}
		return
	}

	req.Res <- defs2.PathDescribeRes{Err: defs2.PathNoStreamAvailableError{PathName: pa.name}}
}

func (pa *PathHandler) doRemovePublisher(req defs2.PathRemovePublisherReq) {
	if pa.source == req.Author {
		pa.executeRemovePublisher()
	}
	close(req.Res)
}

func (pa *PathHandler) doAddPublisher(req defs2.PathAddPublisherReq) {
	if pa.conf.Source != "publisher" {
		req.Res <- defs2.PathAddPublisherRes{
			Err: fmt.Errorf("can't publish to PathHandler '%s' since 'source' is not 'publisher'", pa.name),
		}
		return
	}

	if pa.source != nil {
		if !pa.conf.OverridePublisher {
			req.Res <- defs2.PathAddPublisherRes{Err: fmt.Errorf("someone is already publishing to PathHandler '%s'", pa.name)}
			return
		}

		pa.Log(logger.Info, "closing existing publisher")
		pa.source.(defs2.Publisher).Close()
		pa.executeRemovePublisher()
	}

	pa.source = req.Author
	pa.publisherQuery = req.AccessRequest.Query

	err := pa.setReady(req.Desc, req.GenerateRTPPackets)
	if err != nil {
		pa.source = nil
		req.Res <- defs2.PathAddPublisherRes{Err: err}
		return
	}

	req.Author.Log(logger.Info, "is publishing to PathHandler '%s', %s",
		pa.name,
		defs2.MediasInfo(req.Desc.Medias))

	if pa.conf.HasOnDemandPublisher() && pa.onDemandPublisherState != pathOnDemandStateInitial {
		pa.onDemandPublisherReadyTimer.Stop()
		pa.onDemandPublisherReadyTimer = emptyTimer()
		pa.onDemandPublisherScheduleClose()
	}

	pa.consumeOnHoldRequests()

	req.Res <- defs2.PathAddPublisherRes{
		Path:   pa,
		Stream: pa.stream,
	}
}

func (pa *PathHandler) doAddReader(req defs2.PathAddReaderReq) {
	if pa.stream != nil {
		pa.addReaderPost(req)
		return
	}

	if pa.conf.HasOnDemandStaticSource() {
		if pa.onDemandStaticSourceState == pathOnDemandStateInitial {
			pa.onDemandStaticSourceStart(req.AccessRequest.Query)
		}
		pa.readerAddRequestsOnHold = append(pa.readerAddRequestsOnHold, req)
		return
	}

	if pa.conf.HasOnDemandPublisher() {
		if pa.onDemandPublisherState == pathOnDemandStateInitial {
			pa.onDemandPublisherStart(req.AccessRequest.Query)
		}
		pa.readerAddRequestsOnHold = append(pa.readerAddRequestsOnHold, req)
		return
	}

	req.Res <- defs2.PathAddReaderRes{Err: defs2.PathNoStreamAvailableError{PathName: pa.name}}
}

func (pa *PathHandler) doRemoveReader(req defs2.PathRemoveReaderReq) {
	if _, ok := pa.readers[req.Author]; ok {
		pa.executeRemoveReader(req.Author)
	}
	close(req.Res)

	if len(pa.readers) == 0 {
		if pa.conf.HasOnDemandStaticSource() {
			if pa.onDemandStaticSourceState == pathOnDemandStateReady {
				pa.onDemandStaticSourceScheduleClose()
			}
		} else if pa.conf.HasOnDemandPublisher() {
			if pa.onDemandPublisherState == pathOnDemandStateReady {
				pa.onDemandPublisherScheduleClose()
			}
		}
	}
}

func (pa *PathHandler) doAPIPathsGet(req pathAPIPathsGetReq) {
	req.res <- pathAPIPathsGetRes{
		data: &defs2.APIPath{
			Name:     pa.name,
			ConfName: pa.conf.Name,
			Source: func() *defs2.APIPathSourceOrReader {
				if pa.source == nil {
					return nil
				}
				v := pa.source.APISourceDescribe()
				return &v
			}(),
			Ready: pa.isReady(),
			ReadyTime: func() *time.Time {
				if !pa.isReady() {
					return nil
				}
				v := pa.readyTime
				return &v
			}(),
			Tracks: func() []string {
				if !pa.isReady() {
					return []string{}
				}
				return defs2.MediasToCodecs(pa.stream.Desc.Medias)
			}(),
			BytesReceived: func() uint64 {
				if !pa.isReady() {
					return 0
				}
				return pa.stream.BytesReceived()
			}(),
			BytesSent: func() uint64 {
				if !pa.isReady() {
					return 0
				}
				return pa.stream.BytesSent()
			}(),
			Readers: func() []defs2.APIPathSourceOrReader {
				ret := []defs2.APIPathSourceOrReader{}
				for r := range pa.readers {
					ret = append(ret, r.APIReaderDescribe())
				}
				return ret
			}(),
		},
	}
}

func (pa *PathHandler) SafeConf() *conf2.Path {
	pa.confMutex.RLock()
	defer pa.confMutex.RUnlock()
	return pa.conf
}

func (pa *PathHandler) ExternalCmdEnv() externalcmd.Environment {
	_, port, _ := net.SplitHostPort(pa.rtspAddress)
	env := externalcmd.Environment{
		"MTX_PATH":  pa.name,
		"RTSP_PATH": pa.name, // deprecated
		"RTSP_PORT": port,
	}

	if len(pa.matches) > 1 {
		for i, ma := range pa.matches[1:] {
			env["G"+strconv.FormatInt(int64(i+1), 10)] = ma
		}
	}

	return utils.BuildCmdEnv(env)
}

func (pa *PathHandler) shouldClose() bool {
	return pa.conf.Regexp != nil &&
		pa.source == nil &&
		len(pa.readers) == 0 &&
		len(pa.describeRequestsOnHold) == 0 &&
		len(pa.readerAddRequestsOnHold) == 0
}

func (pa *PathHandler) onDemandStaticSourceStart(query string) {
	pa.source.(*staticsources.Handler).Start(true, query)

	pa.onDemandStaticSourceReadyTimer.Stop()
	pa.onDemandStaticSourceReadyTimer = time.NewTimer(time.Duration(pa.conf.SourceOnDemandStartTimeout))

	pa.onDemandStaticSourceState = pathOnDemandStateWaitingReady
	pa.parent.OnDemandStatic(pa.ctx, pa, query)
}

func (pa *PathHandler) onDemandStaticSourceScheduleClose() {
	pa.onDemandStaticSourceCloseTimer.Stop()
	pa.onDemandStaticSourceCloseTimer = time.NewTimer(time.Duration(pa.conf.SourceOnDemandCloseAfter))

	pa.onDemandStaticSourceState = pathOnDemandStateClosing
}

func (pa *PathHandler) onDemandStaticSourceStop(reason string) {
	if pa.onDemandStaticSourceState == pathOnDemandStateClosing {
		pa.onDemandStaticSourceCloseTimer.Stop()
		pa.onDemandStaticSourceCloseTimer = emptyTimer()
	}

	pa.onDemandStaticSourceState = pathOnDemandStateInitial

	pa.source.(*staticsources.Handler).Stop(reason)
	pa.parent.OnUnDemandStatic(pa.ctx, pa, reason)
}

func (pa *PathHandler) onDemandPublisherStart(query string) {
	pa.onUnDemandHook = hooks.OnDemand(hooks.OnDemandParams{
		Logger:          pa,
		ExternalCmdPool: pa.externalCmdPool,
		Conf:            pa.conf,
		ExternalCmdEnv:  pa.ExternalCmdEnv(),
		Query:           query,
	})

	pa.onDemandPublisherReadyTimer.Stop()
	pa.onDemandPublisherReadyTimer = time.NewTimer(time.Duration(pa.conf.RunOnDemandStartTimeout))

	pa.onDemandPublisherState = pathOnDemandStateWaitingReady
	pa.parent.OnDemand(pa.ctx, pa, query)
}

func (pa *PathHandler) onDemandPublisherScheduleClose() {
	pa.onDemandPublisherCloseTimer.Stop()
	pa.onDemandPublisherCloseTimer = time.NewTimer(time.Duration(pa.conf.RunOnDemandCloseAfter))

	pa.onDemandPublisherState = pathOnDemandStateClosing
}

func (pa *PathHandler) onDemandPublisherStop(reason string) {
	if pa.onDemandPublisherState == pathOnDemandStateClosing {
		pa.onDemandPublisherCloseTimer.Stop()
		pa.onDemandPublisherCloseTimer = emptyTimer()
	}

	pa.onUnDemandHook(reason)
	pa.onUnDemandHook = nil

	pa.onDemandPublisherState = pathOnDemandStateInitial
	pa.parent.OnUnDemand(pa.ctx, pa, reason)
}

func (pa *PathHandler) setReady(desc *description.Session, allocateEncoder bool) error {
	pa.stream = &stream.Stream{
		WriteQueueSize:     pa.writeQueueSize,
		RTPMaxPayloadSize:  pa.rtpMaxPayloadSize,
		Desc:               desc,
		GenerateRTPPackets: allocateEncoder,
		Parent:             pa.source,
	}
	err := pa.stream.Initialize()
	if err != nil {
		return err
	}

	pa.readyTime = time.Now()

	if pa.conf.Record {
		pa.startRecording()
	}

	pa.onNotReadyHook = hooks.OnReady(hooks.OnReadyParams{
		Logger:          pa,
		ExternalCmdPool: pa.externalCmdPool,
		Conf:            pa.conf,
		ExternalCmdEnv:  pa.ExternalCmdEnv(),
		Desc:            pa.source.APISourceDescribe(),
		Query:           pa.publisherQuery,
	})

	pa.parent.pathReady(pa)
	pa.parent.OnReady(pa.ctx, pa)

	return nil
}

func (pa *PathHandler) consumeOnHoldRequests() {
	for _, req := range pa.describeRequestsOnHold {
		req.Res <- defs2.PathDescribeRes{
			Stream: pa.stream,
		}
	}
	pa.describeRequestsOnHold = nil

	for _, req := range pa.readerAddRequestsOnHold {
		pa.addReaderPost(req)
	}
	pa.readerAddRequestsOnHold = nil
}

func (pa *PathHandler) setNotReady() {
	pa.parent.pathNotReady(pa)

	for r := range pa.readers {
		pa.executeRemoveReader(r)
		r.Close()
	}

	pa.onNotReadyHook()

	if pa.recorder != nil {
		pa.recorder.Close()
		pa.recorder = nil
	}

	if pa.stream != nil {
		pa.stream.Close()
		pa.stream = nil
	}

	pa.parent.OnNotReady(pa.ctx, pa)
}

func (pa *PathHandler) startRecording() {
	pa.recorder = &recorder.Recorder{
		PathFormat:      pa.conf.RecordPath,
		Format:          pa.conf.RecordFormat,
		PartDuration:    time.Duration(pa.conf.RecordPartDuration),
		MaxPartSize:     pa.conf.RecordMaxPartSize,
		SegmentDuration: time.Duration(pa.conf.RecordSegmentDuration),
		PathName:        pa.name,
		Stream:          pa.stream,
		OnSegmentCreate: func(segmentPath string) {
			if pa.conf.RunOnRecordSegmentCreate != "" {
				env := pa.ExternalCmdEnv()
				env["MTX_SEGMENT_PATH"] = segmentPath

				pa.Log(logger.Info, "runOnRecordSegmentCreate command launched")
				externalcmd.NewCmd(
					pa.externalCmdPool,
					pa.conf.RunOnRecordSegmentCreate,
					false,
					env,
					nil)
			}
		},
		OnSegmentComplete: func(segmentPath string, segmentDuration time.Duration) {
			if pa.conf.RunOnRecordSegmentComplete != "" {
				env := pa.ExternalCmdEnv()
				env["MTX_SEGMENT_PATH"] = segmentPath
				env["MTX_SEGMENT_DURATION"] = strconv.FormatFloat(segmentDuration.Seconds(), 'f', -1, 64)

				pa.Log(logger.Info, "runOnRecordSegmentComplete command launched")
				externalcmd.NewCmd(
					pa.externalCmdPool,
					pa.conf.RunOnRecordSegmentComplete,
					false,
					env,
					nil)
			}
		},
		Parent: pa,
	}
	pa.recorder.Initialize()
}

func (pa *PathHandler) executeRemoveReader(r defs2.Reader) {
	delete(pa.readers, r)
}

func (pa *PathHandler) executeRemovePublisher() {
	if pa.stream != nil {
		pa.setNotReady()
	}

	pa.source = nil
}

func (pa *PathHandler) addReaderPost(req defs2.PathAddReaderReq) {
	if _, ok := pa.readers[req.Author]; ok {
		req.Res <- defs2.PathAddReaderRes{
			Path:   pa,
			Stream: pa.stream,
		}
		return
	}

	if pa.conf.MaxReaders != 0 && len(pa.readers) >= pa.conf.MaxReaders {
		req.Res <- defs2.PathAddReaderRes{Err: fmt.Errorf("maximum reader count reached")}
		return
	}

	pa.readers[req.Author] = struct{}{}

	if pa.conf.HasOnDemandStaticSource() {
		if pa.onDemandStaticSourceState == pathOnDemandStateClosing {
			pa.onDemandStaticSourceState = pathOnDemandStateReady
			pa.onDemandStaticSourceCloseTimer.Stop()
			pa.onDemandStaticSourceCloseTimer = emptyTimer()
		}
	} else if pa.conf.HasOnDemandPublisher() {
		if pa.onDemandPublisherState == pathOnDemandStateClosing {
			pa.onDemandPublisherState = pathOnDemandStateReady
			pa.onDemandPublisherCloseTimer.Stop()
			pa.onDemandPublisherCloseTimer = emptyTimer()
		}
	}

	req.Res <- defs2.PathAddReaderRes{
		Path:   pa,
		Stream: pa.stream,
	}
}

// reloadConf is called by pathManager.
func (pa *PathHandler) reloadConf(newConf *conf2.Path) {
	select {
	case pa.chReloadConf <- newConf:
	case <-pa.ctx.Done():
	}
}

// StaticSourceHandlerSetReady is called by staticsources.Handler.
func (pa *PathHandler) StaticSourceHandlerSetReady(
	ctx context.Context, req defs2.PathSourceStaticSetReadyReq,
) {
	select {
	case pa.chStaticSourceSetReady <- req:

	case <-pa.ctx.Done():
		req.Res <- defs2.PathSourceStaticSetReadyRes{Err: fmt.Errorf("terminated")}

	// this avoids:
	// - invalid requests sent after the source has been terminated
	// - deadlocks caused by <-Done inside stop()
	case <-ctx.Done():
		req.Res <- defs2.PathSourceStaticSetReadyRes{Err: fmt.Errorf("terminated")}
	}
}

// StaticSourceHandlerSetNotReady is called by staticsources.Handler.
func (pa *PathHandler) StaticSourceHandlerSetNotReady(
	ctx context.Context, req defs2.PathSourceStaticSetNotReadyReq,
) {
	select {
	case pa.chStaticSourceSetNotReady <- req:

	case <-pa.ctx.Done():
		close(req.Res)

	// this avoids:
	// - invalid requests sent after the source has been terminated
	// - deadlocks caused by <-Done inside stop()
	case <-ctx.Done():
		close(req.Res)
	}
}

// describe is called by a reader or publisher through pathManager.
func (pa *PathHandler) describe(req defs2.PathDescribeReq) defs2.PathDescribeRes {
	select {
	case pa.chDescribe <- req:
		return <-req.Res
	case <-pa.ctx.Done():
		return defs2.PathDescribeRes{Err: fmt.Errorf("terminated")}
	}
}

// addPublisher is called by a publisher through pathManager.
func (pa *PathHandler) addPublisher(req defs2.PathAddPublisherReq) (defs2.Path, *stream.Stream, error) {
	select {
	case pa.chAddPublisher <- req:
		res := <-req.Res
		return res.Path, res.Stream, res.Err
	case <-pa.ctx.Done():
		return nil, nil, fmt.Errorf("terminated")
	}
}

// RemovePublisher is called by a publisher.
func (pa *PathHandler) RemovePublisher(req defs2.PathRemovePublisherReq) {
	req.Res = make(chan struct{})
	select {
	case pa.chRemovePublisher <- req:
		<-req.Res
	case <-pa.ctx.Done():
	}
}

// addReader is called by a reader through pathManager.
func (pa *PathHandler) addReader(req defs2.PathAddReaderReq) (defs2.Path, *stream.Stream, error) {
	select {
	case pa.chAddReader <- req:
		res := <-req.Res
		return res.Path, res.Stream, res.Err
	case <-pa.ctx.Done():
		return nil, nil, fmt.Errorf("terminated")
	}
}

// RemoveReader is called by a reader.
func (pa *PathHandler) RemoveReader(req defs2.PathRemoveReaderReq) {
	req.Res = make(chan struct{})
	select {
	case pa.chRemoveReader <- req:
		<-req.Res
	case <-pa.ctx.Done():
	}
}

// APIPathsGet is called by api.
func (pa *PathHandler) APIPathsGet(req pathAPIPathsGetReq) (*defs2.APIPath, error) {
	req.res = make(chan pathAPIPathsGetRes)
	select {
	case pa.chAPIPathsGet <- req:
		res := <-req.res
		return res.data, res.err

	case <-pa.ctx.Done():
		return nil, fmt.Errorf("terminated")
	}
}
