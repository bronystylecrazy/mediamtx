package mtx

import (
	"context"
	"fmt"
	"github.com/bluenviron/mediamtx/pkg/auth"
	conf2 "github.com/bluenviron/mediamtx/pkg/conf"
	defs2 "github.com/bluenviron/mediamtx/pkg/defs"
	"github.com/bluenviron/mediamtx/pkg/externalcmd"
	"github.com/bluenviron/mediamtx/pkg/logger"
	"github.com/bluenviron/mediamtx/pkg/metrics"
	"github.com/bluenviron/mediamtx/pkg/stream"
	"sort"
	"sync"

	// Keep internal packages that we haven't exposed yet
	"github.com/bluenviron/mediamtx/internal/servers/hls"
)

func PathConfCanBeUpdated(oldPathConf *conf2.Path, newPathConf *conf2.Path) bool {
	clone := oldPathConf.Clone()

	clone.Name = newPathConf.Name
	clone.Regexp = newPathConf.Regexp

	clone.Record = newPathConf.Record
	clone.RecordPath = newPathConf.RecordPath
	clone.RecordFormat = newPathConf.RecordFormat
	clone.RecordPartDuration = newPathConf.RecordPartDuration
	clone.RecordMaxPartSize = newPathConf.RecordMaxPartSize
	clone.RecordSegmentDuration = newPathConf.RecordSegmentDuration
	clone.RecordDeleteAfter = newPathConf.RecordDeleteAfter

	clone.RPICameraBrightness = newPathConf.RPICameraBrightness
	clone.RPICameraContrast = newPathConf.RPICameraContrast
	clone.RPICameraSaturation = newPathConf.RPICameraSaturation
	clone.RPICameraSharpness = newPathConf.RPICameraSharpness
	clone.RPICameraExposure = newPathConf.RPICameraExposure
	clone.RPICameraFlickerPeriod = newPathConf.RPICameraFlickerPeriod
	clone.RPICameraAWB = newPathConf.RPICameraAWB
	clone.RPICameraAWBGains = newPathConf.RPICameraAWBGains
	clone.RPICameraDenoise = newPathConf.RPICameraDenoise
	clone.RPICameraShutter = newPathConf.RPICameraShutter
	clone.RPICameraMetering = newPathConf.RPICameraMetering
	clone.RPICameraGain = newPathConf.RPICameraGain
	clone.RPICameraEV = newPathConf.RPICameraEV
	clone.RPICameraFPS = newPathConf.RPICameraFPS
	clone.RPICameraIDRPeriod = newPathConf.RPICameraIDRPeriod
	clone.RPICameraBitrate = newPathConf.RPICameraBitrate

	return newPathConf.Equal(clone)
}

type pathSetHLSServerRes struct {
	readyPaths []defs2.Path
}

type pathSetHLSServerReq struct {
	s   *hls.Server
	res chan pathSetHLSServerRes
}

type pathData struct {
	path     *PathHandler
	ready    bool
	confName string
}

type pathManagerParent interface {
	logger.Writer

	PathHook
}

type pathManager struct {
	logLevel          conf2.LogLevel
	authManager       *auth.Manager
	rtspAddress       string
	readTimeout       conf2.Duration
	writeTimeout      conf2.Duration
	writeQueueSize    int
	rtpMaxPayloadSize int
	pathConfs         map[string]*conf2.Path
	externalCmdPool   *externalcmd.Pool
	metrics           *metrics.Metrics
	parent            pathManagerParent

	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup
	hlsServer *hls.Server
	paths     map[string]*pathData

	// in
	chReloadConf   chan map[string]*conf2.Path
	chSetHLSServer chan pathSetHLSServerReq
	chClosePath    chan *PathHandler
	chPathReady    chan *PathHandler
	chPathNotReady chan *PathHandler
	chFindPathConf chan defs2.PathFindPathConfReq
	chDescribe     chan defs2.PathDescribeReq
	chAddReader    chan defs2.PathAddReaderReq
	chAddPublisher chan defs2.PathAddPublisherReq
	chAPIPathsList chan pathAPIPathsListReq
	chAPIPathsGet  chan pathAPIPathsGetReq
}

func (pm *pathManager) OnDemand(ctx context.Context, handler *PathHandler, s string) {
	pm.parent.OnDemand(ctx, handler, s)
}

func (pm *pathManager) OnUnDemand(ctx context.Context, handler *PathHandler, s string) {
	pm.parent.OnUnDemand(ctx, handler, s)
}

func (pm *pathManager) OnDemandStatic(ctx context.Context, handler *PathHandler, s string) {
	pm.parent.OnDemandStatic(ctx, handler, s)
}

func (pm *pathManager) OnUnDemandStatic(ctx context.Context, handler *PathHandler, s string) {
	pm.parent.OnUnDemandStatic(ctx, handler, s)
}

func (pm *pathManager) OnInit(ctx context.Context, handler *PathHandler) {
	pm.parent.OnInit(ctx, handler)
}

func (pm *pathManager) OnUnInit(ctx context.Context, handler *PathHandler) {
	pm.parent.OnUnInit(ctx, handler)
}

func (pm *pathManager) OnReady(ctx context.Context, handler *PathHandler) {
	pm.parent.OnReady(ctx, handler)
}

func (pm *pathManager) OnNotReady(ctx context.Context, handler *PathHandler) {
	pm.parent.OnNotReady(ctx, handler)
}

func (pm *pathManager) initialize() {
	ctx, ctxCancel := context.WithCancel(context.Background())

	pm.ctx = ctx
	pm.ctxCancel = ctxCancel
	pm.paths = make(map[string]*pathData)
	pm.chReloadConf = make(chan map[string]*conf2.Path)
	pm.chSetHLSServer = make(chan pathSetHLSServerReq)
	pm.chClosePath = make(chan *PathHandler)
	pm.chPathReady = make(chan *PathHandler)
	pm.chPathNotReady = make(chan *PathHandler)
	pm.chFindPathConf = make(chan defs2.PathFindPathConfReq)
	pm.chDescribe = make(chan defs2.PathDescribeReq)
	pm.chAddReader = make(chan defs2.PathAddReaderReq)
	pm.chAddPublisher = make(chan defs2.PathAddPublisherReq)
	pm.chAPIPathsList = make(chan pathAPIPathsListReq)
	pm.chAPIPathsGet = make(chan pathAPIPathsGetReq)

	for _, pathConf := range pm.pathConfs {
		if pathConf.Regexp == nil {
			pm.createPath(pathConf, pathConf.Name, nil)
		}
	}

	pm.Log(logger.Debug, "PathHandler manager created")

	pm.wg.Add(1)
	go pm.run()

	if pm.metrics != nil {
		pm.metrics.SetPathManager(pm)
	}
}

func (pm *pathManager) close() {
	pm.Log(logger.Debug, "PathHandler manager is shutting down")

	if pm.metrics != nil {
		pm.metrics.SetPathManager(nil)
	}

	pm.ctxCancel()
	pm.wg.Wait()
}

// Log implements logger.Writer.
func (pm *pathManager) Log(level logger.Level, format string, args ...interface{}) {
	pm.parent.Log(level, format, args...)
}

func (pm *pathManager) run() {
	defer pm.wg.Done()

outer:
	for {
		select {
		case newPaths := <-pm.chReloadConf:
			pm.doReloadConf(newPaths)

		case req := <-pm.chSetHLSServer:
			readyPaths := pm.doSetHLSServer(req.s)
			req.res <- pathSetHLSServerRes{readyPaths: readyPaths}

		case pa := <-pm.chClosePath:
			pm.doClosePath(pa)

		case pa := <-pm.chPathReady:
			pm.doPathReady(pa)

		case pa := <-pm.chPathNotReady:
			pm.doPathNotReady(pa)

		case req := <-pm.chFindPathConf:
			pm.doFindPathConf(req)

		case req := <-pm.chDescribe:
			pm.doDescribe(req)

		case req := <-pm.chAddReader:
			pm.doAddReader(req)

		case req := <-pm.chAddPublisher:
			pm.doAddPublisher(req)

		case req := <-pm.chAPIPathsList:
			pm.doAPIPathsList(req)

		case req := <-pm.chAPIPathsGet:
			pm.doAPIPathsGet(req)

		case <-pm.ctx.Done():
			break outer
		}
	}

	pm.ctxCancel()
}

func (pm *pathManager) doReloadConf(newPaths map[string]*conf2.Path) {
	confsToRecreate := make(map[string]struct{})
	confsToReload := make(map[string]struct{})

	for confName, pathConf := range pm.pathConfs {
		if newPath, ok := newPaths[confName]; ok {
			if !newPath.Equal(pathConf) {
				if PathConfCanBeUpdated(pathConf, newPath) {
					confsToReload[confName] = struct{}{}
				} else {
					confsToRecreate[confName] = struct{}{}
				}
			}
		}
	}

	// process existing paths
	for pathName, pathData := range pm.paths {
		path := pathData.path
		newPathConf, _, err := conf2.FindPathConf(newPaths, pathName)
		// PathHandler does not have a config anymore: delete it
		if err != nil {
			pm.removeAndClosePath(path)
			continue
		}

		// PathHandler now belongs to a different config
		if newPathConf.Name != pathData.confName {
			// PathHandler config can be hot reloaded
			oldPathConf := pm.pathConfs[pathData.confName]
			if PathConfCanBeUpdated(oldPathConf, newPathConf) {
				pm.paths[path.name].confName = newPathConf.Name
				go path.reloadConf(newPathConf)
				continue
			}

			// Configuration cannot be hot reloaded: delete the PathHandler
			pm.removeAndClosePath(path)
			continue
		}

		// PathHandler configuration has changed and cannot be hot reloaded: delete PathHandler
		if _, ok := confsToRecreate[newPathConf.Name]; ok {
			pm.removeAndClosePath(path)
			continue
		}

		// PathHandler configuration has changed but can be hot reloaded: reload it
		if _, ok := confsToReload[newPathConf.Name]; ok {
			go path.reloadConf(newPathConf)
		}
	}

	pm.pathConfs = newPaths

	// create new static paths
	for pathConfName, pathConf := range newPaths {
		if pathConf.Regexp == nil {
			if _, ok := pm.paths[pathConfName]; !ok {
				pm.createPath(pathConf, pathConfName, nil)
			}
		}
	}
}

func (pm *pathManager) removeAndClosePath(path *PathHandler) {
	pm.removePath(path)
	path.close()
	path.wait() // avoid conflicts between sources
}

func (pm *pathManager) doSetHLSServer(m *hls.Server) []defs2.Path {
	pm.hlsServer = m

	var ret []defs2.Path

	for _, pd := range pm.paths {
		if pd.ready {
			ret = append(ret, pd.path)
		}
	}

	return ret
}

func (pm *pathManager) doClosePath(pa *PathHandler) {
	if pd, ok := pm.paths[pa.name]; !ok || pd.path != pa {
		return
	}
	pm.removePath(pa)
}

func (pm *pathManager) doPathReady(pa *PathHandler) {
	if pd, ok := pm.paths[pa.name]; !ok || pd.path != pa {
		return
	}

	pm.paths[pa.name].ready = true

	if pm.hlsServer != nil {
		pm.hlsServer.PathReady(pa)
	}
}

func (pm *pathManager) doPathNotReady(pa *PathHandler) {
	if pd, ok := pm.paths[pa.name]; !ok || pd.path != pa {
		return
	}

	pm.paths[pa.name].ready = false

	if pm.hlsServer != nil {
		pm.hlsServer.PathNotReady(pa)
	}
}

func (pm *pathManager) doFindPathConf(req defs2.PathFindPathConfReq) {
	pathConf, _, err := conf2.FindPathConf(pm.pathConfs, req.AccessRequest.Name)
	if err != nil {
		req.Res <- defs2.PathFindPathConfRes{Err: err}
		return
	}

	err2 := pm.authManager.Authenticate(req.AccessRequest.ToAuthRequest())
	if err2 != nil {
		req.Res <- defs2.PathFindPathConfRes{Err: err2}
		return
	}

	req.Res <- defs2.PathFindPathConfRes{Conf: pathConf}
}

func (pm *pathManager) doDescribe(req defs2.PathDescribeReq) {
	pathConf, pathMatches, err := conf2.FindPathConf(pm.pathConfs, req.AccessRequest.Name)
	if err != nil {
		req.Res <- defs2.PathDescribeRes{Err: err}
		return
	}

	err2 := pm.authManager.Authenticate(req.AccessRequest.ToAuthRequest())
	if err2 != nil {
		req.Res <- defs2.PathDescribeRes{Err: err2}
		return
	}

	// create PathHandler if it doesn't exist
	if _, ok := pm.paths[req.AccessRequest.Name]; !ok {
		pm.createPath(pathConf, req.AccessRequest.Name, pathMatches)
	}

	pd := pm.paths[req.AccessRequest.Name]
	req.Res <- defs2.PathDescribeRes{Path: pd.path}
}

func (pm *pathManager) doAddReader(req defs2.PathAddReaderReq) {
	pathConf, pathMatches, err := conf2.FindPathConf(pm.pathConfs, req.AccessRequest.Name)
	if err != nil {
		req.Res <- defs2.PathAddReaderRes{Err: err}
		return
	}

	if !req.AccessRequest.SkipAuth {
		err2 := pm.authManager.Authenticate(req.AccessRequest.ToAuthRequest())
		if err2 != nil {
			req.Res <- defs2.PathAddReaderRes{Err: err2}
			return
		}
	}

	// create PathHandler if it doesn't exist
	if _, ok := pm.paths[req.AccessRequest.Name]; !ok {
		pm.createPath(pathConf, req.AccessRequest.Name, pathMatches)
	}

	pd := pm.paths[req.AccessRequest.Name]
	req.Res <- defs2.PathAddReaderRes{Path: pd.path}
}

func (pm *pathManager) doAddPublisher(req defs2.PathAddPublisherReq) {
	pathConf, pathMatches, err := conf2.FindPathConf(pm.pathConfs, req.AccessRequest.Name)
	if err != nil {
		req.Res <- defs2.PathAddPublisherRes{Err: err}
		return
	}

	if req.ConfToCompare != nil && !pathConf.Equal(req.ConfToCompare) {
		req.Res <- defs2.PathAddPublisherRes{Err: fmt.Errorf("configuration has changed")}
		return
	}

	if !req.AccessRequest.SkipAuth {
		err2 := pm.authManager.Authenticate(req.AccessRequest.ToAuthRequest())
		if err2 != nil {
			req.Res <- defs2.PathAddPublisherRes{Err: err2}
			return
		}
	}

	// create PathHandler if it doesn't exist
	if _, ok := pm.paths[req.AccessRequest.Name]; !ok {
		pm.createPath(pathConf, req.AccessRequest.Name, pathMatches)
	}

	pd := pm.paths[req.AccessRequest.Name]
	req.Res <- defs2.PathAddPublisherRes{Path: pd.path}
}

func (pm *pathManager) doAPIPathsList(req pathAPIPathsListReq) {
	paths := make(map[string]*PathHandler)

	for name, pd := range pm.paths {
		paths[name] = pd.path
	}

	req.res <- pathAPIPathsListRes{paths: paths}
}

func (pm *pathManager) doAPIPathsGet(req pathAPIPathsGetReq) {
	pd, ok := pm.paths[req.name]
	if !ok {
		req.res <- pathAPIPathsGetRes{err: conf2.ErrPathNotFound}
		return
	}

	req.res <- pathAPIPathsGetRes{path: pd.path}
}

func (pm *pathManager) createPath(
	pathConf *conf2.Path,
	name string,
	matches []string,
) {
	pa := &PathHandler{
		parentCtx:         pm.ctx,
		logLevel:          pm.logLevel,
		rtspAddress:       pm.rtspAddress,
		readTimeout:       pm.readTimeout,
		writeTimeout:      pm.writeTimeout,
		writeQueueSize:    pm.writeQueueSize,
		rtpMaxPayloadSize: pm.rtpMaxPayloadSize,
		conf:              pathConf,
		name:              name,
		matches:           matches,
		wg:                &pm.wg,
		externalCmdPool:   pm.externalCmdPool,
		parent:            pm,
	}
	pa.initialize()

	pm.paths[name] = &pathData{
		path:     pa,
		confName: pathConf.Name,
	}
}

func (pm *pathManager) removePath(pa *PathHandler) {
	delete(pm.paths, pa.name)
}

// ReloadPathConfs is called by core.
func (pm *pathManager) ReloadPathConfs(pathConfs map[string]*conf2.Path) {
	select {
	case pm.chReloadConf <- pathConfs:
	case <-pm.ctx.Done():
	}
}

// pathReady is called by PathHandler.
func (pm *pathManager) pathReady(pa *PathHandler) {
	select {
	case pm.chPathReady <- pa:
	case <-pm.ctx.Done():
	case <-pa.ctx.Done(): // in case PathManager is blocked by PathHandler.wait()
	}
}

// pathNotReady is called by PathHandler.
func (pm *pathManager) pathNotReady(pa *PathHandler) {
	select {
	case pm.chPathNotReady <- pa:
	case <-pm.ctx.Done():
	case <-pa.ctx.Done(): // in case PathManager is blocked by PathHandler.wait()
	}
}

// closePath is called by PathHandler.
func (pm *pathManager) closePath(pa *PathHandler) {
	select {
	case pm.chClosePath <- pa:
	case <-pm.ctx.Done():
	case <-pa.ctx.Done(): // in case PathManager is blocked by PathHandler.wait()
	}
}

// FindPathConf is called by a reader or publisher.
func (pm *pathManager) FindPathConf(req defs2.PathFindPathConfReq) (*conf2.Path, error) {
	req.Res = make(chan defs2.PathFindPathConfRes)
	select {
	case pm.chFindPathConf <- req:
		res := <-req.Res
		return res.Conf, res.Err

	case <-pm.ctx.Done():
		return nil, fmt.Errorf("terminated")
	}
}

// Describe is called by a reader or publisher.
func (pm *pathManager) Describe(req defs2.PathDescribeReq) defs2.PathDescribeRes {
	req.Res = make(chan defs2.PathDescribeRes)
	select {
	case pm.chDescribe <- req:
		res1 := <-req.Res
		if res1.Err != nil {
			return res1
		}

		res2 := res1.Path.(*PathHandler).describe(req)
		if res2.Err != nil {
			return res2
		}

		res2.Path = res1.Path
		return res2

	case <-pm.ctx.Done():
		return defs2.PathDescribeRes{Err: fmt.Errorf("terminated")}
	}
}

// AddPublisher is called by a publisher.
func (pm *pathManager) AddPublisher(req defs2.PathAddPublisherReq) (defs2.Path, *stream.Stream, error) {
	req.Res = make(chan defs2.PathAddPublisherRes)
	select {
	case pm.chAddPublisher <- req:
		res := <-req.Res
		if res.Err != nil {
			return nil, nil, res.Err
		}

		return res.Path.(*PathHandler).addPublisher(req)

	case <-pm.ctx.Done():
		return nil, nil, fmt.Errorf("terminated")
	}
}

// AddReader is called by a reader.
func (pm *pathManager) AddReader(req defs2.PathAddReaderReq) (defs2.Path, *stream.Stream, error) {
	req.Res = make(chan defs2.PathAddReaderRes)
	select {
	case pm.chAddReader <- req:
		res := <-req.Res
		if res.Err != nil {
			return nil, nil, res.Err
		}

		return res.Path.(*PathHandler).addReader(req)

	case <-pm.ctx.Done():
		return nil, nil, fmt.Errorf("terminated")
	}
}

// SetHLSServer is called by hls.Server.
func (pm *pathManager) SetHLSServer(s *hls.Server) []defs2.Path {
	req := pathSetHLSServerReq{
		s:   s,
		res: make(chan pathSetHLSServerRes),
	}

	select {
	case pm.chSetHLSServer <- req:
		res := <-req.res
		return res.readyPaths

	case <-pm.ctx.Done():
		return nil
	}
}

// APIPathsList is called by api.
func (pm *pathManager) APIPathsList() (*defs2.APIPathList, error) {
	req := pathAPIPathsListReq{
		res: make(chan pathAPIPathsListRes),
	}

	select {
	case pm.chAPIPathsList <- req:
		res := <-req.res

		res.data = &defs2.APIPathList{
			Items: []*defs2.APIPath{},
		}

		for _, pa := range res.paths {
			item, err := pa.APIPathsGet(pathAPIPathsGetReq{})
			if err == nil {
				res.data.Items = append(res.data.Items, item)
			}
		}

		sort.Slice(res.data.Items, func(i, j int) bool {
			return res.data.Items[i].Name < res.data.Items[j].Name
		})

		return res.data, nil

	case <-pm.ctx.Done():
		return nil, fmt.Errorf("terminated")
	}
}

// APIPathsGet is called by api.
func (pm *pathManager) APIPathsGet(name string) (*defs2.APIPath, error) {
	req := pathAPIPathsGetReq{
		name: name,
		res:  make(chan pathAPIPathsGetRes),
	}

	select {
	case pm.chAPIPathsGet <- req:
		res := <-req.res
		if res.err != nil {
			return nil, res.err
		}

		data, err := res.path.APIPathsGet(req)
		return data, err

	case <-pm.ctx.Done():
		return nil, fmt.Errorf("terminated")
	}
}
