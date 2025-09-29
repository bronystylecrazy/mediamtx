// Package core contains the main struct of the software.
package mediamtx

import (
	"context"
	_ "embed"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/gin-gonic/gin"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/auth"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/externalcmd"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/logger"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/metrics"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/playback"

	// Keep internal packages that we haven't exposed yet
	"github.com/bluenviron/mediamtx/internal/confwatcher"
	"github.com/bluenviron/mediamtx/internal/pprof"
	"github.com/bluenviron/mediamtx/internal/recordcleaner"
	"github.com/bluenviron/mediamtx/internal/rlimit"
	"github.com/bluenviron/mediamtx/internal/servers/hls"
	"github.com/bluenviron/mediamtx/internal/servers/rtmp"
	"github.com/bluenviron/mediamtx/internal/servers/rtsp"
	"github.com/bluenviron/mediamtx/internal/servers/srt"
	"github.com/bluenviron/mediamtx/internal/servers/webrtc"
)

var defaultConfPaths = []string{
	"rtsp-simple-server.yml",
	"mediamtx.yml",
	"/usr/local/etc/mediamtx.yml",
	"/usr/etc/mediamtx.yml",
	"/etc/mediamtx/mediamtx.yml",
}

var cli struct {
	Version  bool   `help:"print version"`
	Confpath string `arg:"" default:""`
}

func atLeastOneRecordDeleteAfter(pathConfs map[string]*conf.Path) bool {
	for _, e := range pathConfs {
		if e.RecordDeleteAfter != 0 {
			return true
		}
	}
	return false
}

func getRTPMaxPayloadSize(udpMaxPayloadSize int, rtspEncryption conf.Encryption) int {
	// UDP max payload size - 12 (RTP header)
	v := udpMaxPayloadSize - 12

	// 10 (SRTP HMAC SHA1 authentication tag)
	if rtspEncryption == conf.EncryptionOptional || rtspEncryption == conf.EncryptionStrict {
		v -= 10
	}

	return v
}

// Core is an instance of MediaMTX.
type Core struct {
	ctx             context.Context
	ctxCancel       func()
	ConfPath        string
	Conf            *conf.Conf
	Logger          *logger.Logger
	ExternalCmdPool *externalcmd.Pool
	AuthManager     *auth.Manager
	Metrics         *metrics.Metrics
	Pprof           *pprof.PPROF
	recordCleaner   *recordcleaner.Cleaner
	playbackServer  *playback.Server
	PathManager     *pathManager
	RtspServer      *rtsp.Server
	RtspsServer     *rtsp.Server
	RtmpServer      *rtmp.Server
	RtmpsServer     *rtmp.Server
	HlsServer       *hls.Server
	WebRTCServer    *webrtc.Server
	SrtServer       *srt.Server
	// api             *api.API
	ConfWatcher *confwatcher.ConfWatcher

	// Options
	opts Options

	// in
	ChAPIConfigSet chan *conf.Conf

	// out
	Done chan struct{}
}

type LogFunc = func(level LogLevel, format string, args ...interface{})

type Options struct {
	LogFunc  LogFunc
	PathHook PathHook
}

// New allocates a Core.
func New(opts Options) (*Core, error) {

	ctx, ctxCancel := context.WithCancel(context.Background())

	p := &Core{
		ctx:            ctx,
		ctxCancel:      ctxCancel,
		ChAPIConfigSet: make(chan *conf.Conf),
		Done:           make(chan struct{}),
		opts:           opts,
	}

	tempLogger, _ := logger.New(logger.Warn, []logger.Destination{logger.DestinationStdout}, "", "")

	cfg, confPath, err := conf.Load(cli.Confpath, defaultConfPaths, tempLogger)
	if err != nil {
		return nil, err
	}

	p.Conf = cfg
	p.ConfPath = confPath

	return p, nil
}

func (p *Core) Run(ctx context.Context) error {
	err := p.CreateResources(true)
	if err != nil {
		if p.Logger != nil {
			p.Log(logger.Error, "%s", err)
		}
		p.closeResources(nil, false)
		return err
	}

	go p.run()
	return nil
}

// Close closes Core and waits for all goroutines to return.
func (p *Core) Close(ctx context.Context) error {
	p.ctxCancel()
	<-p.Done
	return nil
}

// Wait waits for the Core to exit.
func (p *Core) Wait() {
	<-p.Done
}

// Log implements logger.Writer.
func (p *Core) Log(level LogLevel, format string, args ...interface{}) {
	if p.opts.LogFunc == nil {
		p.Logger.Log(level, format, args...)
		return
	}
	p.opts.LogFunc(level, format, args...)
}

func (p *Core) run() {
	defer close(p.Done)

	confChanged := func() chan struct{} {
		if p.ConfWatcher != nil {
			return p.ConfWatcher.Watch()
		}
		return make(chan struct{})
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	if runtime.GOOS == "linux" {
		signal.Notify(interrupt, syscall.SIGTERM)
	}

outer:
	for {
		select {
		case <-confChanged:
			p.Log(logger.Info, "reloading configuration (file changed)")

			newConf, _, err := conf.Load(p.ConfPath, nil, p.Logger)
			if err != nil {
				p.Log(logger.Error, "%s", err)
				break outer
			}

			err = p.reloadConf(newConf, false)
			if err != nil {
				p.Log(logger.Error, "%s", err)
				break outer
			}

		case newConf := <-p.ChAPIConfigSet:
			p.Log(logger.Info, "reloading configuration (API request)")

			err := p.reloadConf(newConf, true)
			if err != nil {
				p.Log(logger.Error, "%s", err)
				break outer
			}

		case <-interrupt:
			p.Log(logger.Info, "shutting down gracefully")
			break outer

		case <-p.ctx.Done():
			break outer
		}
	}

	p.ctxCancel()

	p.closeResources(nil, false)
}

func (p *Core) CreateResources(initial bool) error {
	var err error

	if p.Logger == nil {
		p.Logger, err = logger.New(
			logger.Level(p.Conf.LogLevel),
			p.Conf.LogDestinations,
			p.Conf.LogFile,
			p.Conf.SysLogPrefix,
		)
		if err != nil {
			return err
		}
	}

	if initial {
		if p.ConfPath != "" {
			a, _ := filepath.Abs(p.ConfPath)
			p.Log(logger.Info, "configuration loaded from %s", a)
		} else {
			list := make([]string, len(defaultConfPaths))
			for i, pa := range defaultConfPaths {
				a, _ := filepath.Abs(pa)
				list[i] = a
			}

			p.Log(logger.Warn,
				"configuration file not found (looked in %s), using an empty configuration",
				strings.Join(list, ", "))
		}

		// on Linux, try to raise the number of file descriptors that can be opened
		// to allow the maximum possible number of clients.
		rlimit.Raise() //nolint:errcheck

		gin.SetMode(gin.ReleaseMode)

		p.ExternalCmdPool = &externalcmd.Pool{}
		p.ExternalCmdPool.Initialize()
	}

	if p.AuthManager == nil {
		p.AuthManager = &auth.Manager{
			Method:             p.Conf.AuthMethod,
			InternalUsers:      p.Conf.AuthInternalUsers,
			HTTPAddress:        p.Conf.AuthHTTPAddress,
			HTTPExclude:        p.Conf.AuthHTTPExclude,
			JWTJWKS:            p.Conf.AuthJWTJWKS,
			JWTJWKSFingerprint: p.Conf.AuthJWTJWKSFingerprint,
			JWTClaimKey:        p.Conf.AuthJWTClaimKey,
			JWTExclude:         p.Conf.AuthJWTExclude,
			JWTInHTTPQuery:     p.Conf.AuthJWTInHTTPQuery,
			ReadTimeout:        time.Duration(p.Conf.ReadTimeout),
		}
	}

	if p.Conf.Metrics &&
		p.Metrics == nil {
		i := &metrics.Metrics{
			Address:        p.Conf.MetricsAddress,
			Encryption:     p.Conf.MetricsEncryption,
			ServerKey:      p.Conf.MetricsServerKey,
			ServerCert:     p.Conf.MetricsServerCert,
			AllowOrigin:    p.Conf.MetricsAllowOrigin,
			TrustedProxies: p.Conf.MetricsTrustedProxies,
			ReadTimeout:    p.Conf.ReadTimeout,
			AuthManager:    p.AuthManager,
			Parent:         p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.Metrics = i
	}

	if p.Conf.PPROF &&
		p.Pprof == nil {
		i := &pprof.PPROF{
			Address:        p.Conf.PPROFAddress,
			Encryption:     p.Conf.PPROFEncryption,
			ServerKey:      p.Conf.PPROFServerKey,
			ServerCert:     p.Conf.PPROFServerCert,
			AllowOrigin:    p.Conf.PPROFAllowOrigin,
			TrustedProxies: p.Conf.PPROFTrustedProxies,
			ReadTimeout:    p.Conf.ReadTimeout,
			AuthManager:    p.AuthManager,
			Parent:         p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.Pprof = i
	}

	if p.recordCleaner == nil &&
		atLeastOneRecordDeleteAfter(p.Conf.Paths) {
		p.recordCleaner = &recordcleaner.Cleaner{
			PathConfs: p.Conf.Paths,
			Parent:    p,
		}
		p.recordCleaner.Initialize()
	}

	if p.Conf.Playback &&
		p.playbackServer == nil {
		i := &playback.Server{
			Address:        p.Conf.PlaybackAddress,
			Encryption:     p.Conf.PlaybackEncryption,
			ServerKey:      p.Conf.PlaybackServerKey,
			ServerCert:     p.Conf.PlaybackServerCert,
			AllowOrigin:    p.Conf.PlaybackAllowOrigin,
			TrustedProxies: p.Conf.PlaybackTrustedProxies,
			ReadTimeout:    p.Conf.ReadTimeout,
			PathConfs:      p.Conf.Paths,
			AuthManager:    p.AuthManager,
			Parent:         p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.playbackServer = i
	}

	if p.PathManager == nil {
		rtpMaxPayloadSize := getRTPMaxPayloadSize(p.Conf.UDPMaxPayloadSize, p.Conf.RTSPEncryption)

		p.PathManager = &pathManager{
			logLevel:          p.Conf.LogLevel,
			authManager:       p.AuthManager,
			rtspAddress:       p.Conf.RTSPAddress,
			readTimeout:       p.Conf.ReadTimeout,
			writeTimeout:      p.Conf.WriteTimeout,
			writeQueueSize:    p.Conf.WriteQueueSize,
			rtpMaxPayloadSize: rtpMaxPayloadSize,
			pathConfs:         p.Conf.Paths,
			externalCmdPool:   p.ExternalCmdPool,
			metrics:           p.Metrics,
			parent:            p,
		}
		p.PathManager.initialize()
	}

	if p.Conf.RTSP &&
		(p.Conf.RTSPEncryption == conf.EncryptionNo ||
			p.Conf.RTSPEncryption == conf.EncryptionOptional) &&
		p.RtspServer == nil {
		_, useUDP := p.Conf.RTSPTransports[gortsplib.ProtocolUDP]
		_, useMulticast := p.Conf.RTSPTransports[gortsplib.ProtocolUDPMulticast]

		i := &rtsp.Server{
			Address:             p.Conf.RTSPAddress,
			AuthMethods:         p.Conf.RTSPAuthMethods,
			UDPReadBufferSize:   p.Conf.RTSPUDPReadBufferSize,
			ReadTimeout:         p.Conf.ReadTimeout,
			WriteTimeout:        p.Conf.WriteTimeout,
			WriteQueueSize:      p.Conf.WriteQueueSize,
			UseUDP:              useUDP,
			UseMulticast:        useMulticast,
			RTPAddress:          p.Conf.RTPAddress,
			RTCPAddress:         p.Conf.RTCPAddress,
			MulticastIPRange:    p.Conf.MulticastIPRange,
			MulticastRTPPort:    p.Conf.MulticastRTPPort,
			MulticastRTCPPort:   p.Conf.MulticastRTCPPort,
			IsTLS:               false,
			ServerCert:          "",
			ServerKey:           "",
			RTSPAddress:         p.Conf.RTSPAddress,
			Transports:          p.Conf.RTSPTransports,
			RunOnConnect:        p.Conf.RunOnConnect,
			RunOnConnectRestart: p.Conf.RunOnConnectRestart,
			RunOnDisconnect:     p.Conf.RunOnDisconnect,
			ExternalCmdPool:     p.ExternalCmdPool,
			Metrics:             p.Metrics,
			PathManager:         p.PathManager,
			Parent:              p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.RtspServer = i
	}

	if p.Conf.RTSP &&
		(p.Conf.RTSPEncryption == conf.EncryptionStrict ||
			p.Conf.RTSPEncryption == conf.EncryptionOptional) &&
		p.RtspsServer == nil {
		_, useUDP := p.Conf.RTSPTransports[gortsplib.ProtocolUDP]
		_, useMulticast := p.Conf.RTSPTransports[gortsplib.ProtocolUDPMulticast]

		i := &rtsp.Server{
			Address:             p.Conf.RTSPSAddress,
			AuthMethods:         p.Conf.RTSPAuthMethods,
			UDPReadBufferSize:   p.Conf.RTSPUDPReadBufferSize,
			ReadTimeout:         p.Conf.ReadTimeout,
			WriteTimeout:        p.Conf.WriteTimeout,
			WriteQueueSize:      p.Conf.WriteQueueSize,
			UseUDP:              useUDP,
			UseMulticast:        useMulticast,
			RTPAddress:          p.Conf.SRTPAddress,
			RTCPAddress:         p.Conf.SRTCPAddress,
			MulticastIPRange:    p.Conf.MulticastIPRange,
			MulticastRTPPort:    p.Conf.MulticastSRTPPort,
			MulticastRTCPPort:   p.Conf.MulticastSRTCPPort,
			IsTLS:               true,
			ServerCert:          p.Conf.RTSPServerCert,
			ServerKey:           p.Conf.RTSPServerKey,
			RTSPAddress:         p.Conf.RTSPAddress,
			Transports:          p.Conf.RTSPTransports,
			RunOnConnect:        p.Conf.RunOnConnect,
			RunOnConnectRestart: p.Conf.RunOnConnectRestart,
			RunOnDisconnect:     p.Conf.RunOnDisconnect,
			ExternalCmdPool:     p.ExternalCmdPool,
			Metrics:             p.Metrics,
			PathManager:         p.PathManager,
			Parent:              p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.RtspsServer = i
	}

	if p.Conf.RTMP &&
		(p.Conf.RTMPEncryption == conf.EncryptionNo ||
			p.Conf.RTMPEncryption == conf.EncryptionOptional) &&
		p.RtmpServer == nil {
		i := &rtmp.Server{
			Address:             p.Conf.RTMPAddress,
			ReadTimeout:         p.Conf.ReadTimeout,
			WriteTimeout:        p.Conf.WriteTimeout,
			IsTLS:               false,
			ServerCert:          "",
			ServerKey:           "",
			RTSPAddress:         p.Conf.RTSPAddress,
			RunOnConnect:        p.Conf.RunOnConnect,
			RunOnConnectRestart: p.Conf.RunOnConnectRestart,
			RunOnDisconnect:     p.Conf.RunOnDisconnect,
			ExternalCmdPool:     p.ExternalCmdPool,
			Metrics:             p.Metrics,
			PathManager:         p.PathManager,
			Parent:              p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.RtmpServer = i
	}

	if p.Conf.RTMP &&
		(p.Conf.RTMPEncryption == conf.EncryptionStrict ||
			p.Conf.RTMPEncryption == conf.EncryptionOptional) &&
		p.RtmpsServer == nil {
		i := &rtmp.Server{
			Address:             p.Conf.RTMPSAddress,
			ReadTimeout:         p.Conf.ReadTimeout,
			WriteTimeout:        p.Conf.WriteTimeout,
			IsTLS:               true,
			ServerCert:          p.Conf.RTMPServerCert,
			ServerKey:           p.Conf.RTMPServerKey,
			RTSPAddress:         p.Conf.RTSPAddress,
			RunOnConnect:        p.Conf.RunOnConnect,
			RunOnConnectRestart: p.Conf.RunOnConnectRestart,
			RunOnDisconnect:     p.Conf.RunOnDisconnect,
			ExternalCmdPool:     p.ExternalCmdPool,
			Metrics:             p.Metrics,
			PathManager:         p.PathManager,
			Parent:              p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.RtmpsServer = i
	}

	if p.Conf.HLS &&
		p.HlsServer == nil {
		i := &hls.Server{
			Address:         p.Conf.HLSAddress,
			Encryption:      p.Conf.HLSEncryption,
			ServerKey:       p.Conf.HLSServerKey,
			ServerCert:      p.Conf.HLSServerCert,
			AllowOrigin:     p.Conf.HLSAllowOrigin,
			TrustedProxies:  p.Conf.HLSTrustedProxies,
			AlwaysRemux:     p.Conf.HLSAlwaysRemux,
			Variant:         p.Conf.HLSVariant,
			SegmentCount:    p.Conf.HLSSegmentCount,
			SegmentDuration: p.Conf.HLSSegmentDuration,
			PartDuration:    p.Conf.HLSPartDuration,
			SegmentMaxSize:  p.Conf.HLSSegmentMaxSize,
			Directory:       p.Conf.HLSDirectory,
			ReadTimeout:     p.Conf.ReadTimeout,
			MuxerCloseAfter: p.Conf.HLSMuxerCloseAfter,
			Metrics:         p.Metrics,
			PathManager:     p.PathManager,
			Parent:          p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.HlsServer = i
	}

	if p.Conf.WebRTC &&
		p.WebRTCServer == nil {
		i := &webrtc.Server{
			Address:               p.Conf.WebRTCAddress,
			Encryption:            p.Conf.WebRTCEncryption,
			ServerKey:             p.Conf.WebRTCServerKey,
			ServerCert:            p.Conf.WebRTCServerCert,
			AllowOrigin:           p.Conf.WebRTCAllowOrigin,
			TrustedProxies:        p.Conf.WebRTCTrustedProxies,
			ReadTimeout:           p.Conf.ReadTimeout,
			LocalUDPAddress:       p.Conf.WebRTCLocalUDPAddress,
			LocalTCPAddress:       p.Conf.WebRTCLocalTCPAddress,
			IPsFromInterfaces:     p.Conf.WebRTCIPsFromInterfaces,
			IPsFromInterfacesList: p.Conf.WebRTCIPsFromInterfacesList,
			AdditionalHosts:       p.Conf.WebRTCAdditionalHosts,
			ICEServers:            p.Conf.WebRTCICEServers2,
			HandshakeTimeout:      p.Conf.WebRTCHandshakeTimeout,
			STUNGatherTimeout:     p.Conf.WebRTCSTUNGatherTimeout,
			TrackGatherTimeout:    p.Conf.WebRTCTrackGatherTimeout,
			ExternalCmdPool:       p.ExternalCmdPool,
			Metrics:               p.Metrics,
			PathManager:           p.PathManager,
			Parent:                p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.WebRTCServer = i
	}

	if p.Conf.SRT &&
		p.SrtServer == nil {
		i := &srt.Server{
			Address:             p.Conf.SRTAddress,
			RTSPAddress:         p.Conf.RTSPAddress,
			ReadTimeout:         p.Conf.ReadTimeout,
			WriteTimeout:        p.Conf.WriteTimeout,
			UDPMaxPayloadSize:   p.Conf.UDPMaxPayloadSize,
			RunOnConnect:        p.Conf.RunOnConnect,
			RunOnConnectRestart: p.Conf.RunOnConnectRestart,
			RunOnDisconnect:     p.Conf.RunOnDisconnect,
			ExternalCmdPool:     p.ExternalCmdPool,
			Metrics:             p.Metrics,
			PathManager:         p.PathManager,
			Parent:              p,
		}
		err = i.Initialize()
		if err != nil {
			return err
		}
		p.SrtServer = i
	}

	// if p.Conf.API &&
	// 	p.api == nil {
	// 	i := &api.API{
	// 		Address:        p.Conf.APIAddress,
	// 		Encryption:     p.Conf.APIEncryption,
	// 		ServerKey:      p.Conf.APIServerKey,
	// 		ServerCert:     p.Conf.APIServerCert,
	// 		AllowOrigin:    p.Conf.APIAllowOrigin,
	// 		TrustedProxies: p.Conf.APITrustedProxies,
	// 		ReadTimeout:    p.Conf.ReadTimeout,
	// 		Conf:           p.Conf,
	// 		AuthManager:    p.AuthManager,
	// 		PathManager:    p.PathManager,
	// 		RTSPServer:     p.RtspServer,
	// 		RTSPSServer:    p.RtspsServer,
	// 		RTMPServer:     p.RtmpServer,
	// 		RTMPSServer:    p.RtmpsServer,
	// 		HLSServer:      p.HlsServer,
	// 		WebRTCServer:   p.WebRTCServer,
	// 		SRTServer:      p.SrtServer,
	// 		Parent:         p,
	// 	}
	// 	err = i.Initialize()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	p.api = i
	// }

	if initial && p.ConfPath != "" {
		cf := &confwatcher.ConfWatcher{FilePath: p.ConfPath}
		err = cf.Initialize()
		if err != nil {
			return err
		}
		p.ConfWatcher = cf
	}

	return nil
}

func (p *Core) closeResources(newConf *conf.Conf, calledByAPI bool) {
	closeLogger := newConf == nil ||
		newConf.LogLevel != p.Conf.LogLevel ||
		!reflect.DeepEqual(newConf.LogDestinations, p.Conf.LogDestinations) ||
		newConf.LogFile != p.Conf.LogFile ||
		newConf.SysLogPrefix != p.Conf.SysLogPrefix

	closeAuthManager := newConf == nil ||
		newConf.AuthMethod != p.Conf.AuthMethod ||
		newConf.AuthHTTPAddress != p.Conf.AuthHTTPAddress ||
		!reflect.DeepEqual(newConf.AuthHTTPExclude, p.Conf.AuthHTTPExclude) ||
		newConf.AuthJWTJWKS != p.Conf.AuthJWTJWKS ||
		newConf.AuthJWTJWKSFingerprint != p.Conf.AuthJWTJWKSFingerprint ||
		newConf.AuthJWTClaimKey != p.Conf.AuthJWTClaimKey ||
		!reflect.DeepEqual(newConf.AuthJWTExclude, p.Conf.AuthJWTExclude) ||
		newConf.AuthJWTInHTTPQuery != p.Conf.AuthJWTInHTTPQuery ||
		newConf.ReadTimeout != p.Conf.ReadTimeout
	if !closeAuthManager && !reflect.DeepEqual(newConf.AuthInternalUsers, p.Conf.AuthInternalUsers) {
		p.AuthManager.ReloadInternalUsers(newConf.AuthInternalUsers)
	}

	closeMetrics := newConf == nil ||
		newConf.Metrics != p.Conf.Metrics ||
		newConf.MetricsAddress != p.Conf.MetricsAddress ||
		newConf.MetricsEncryption != p.Conf.MetricsEncryption ||
		newConf.MetricsServerKey != p.Conf.MetricsServerKey ||
		newConf.MetricsServerCert != p.Conf.MetricsServerCert ||
		newConf.MetricsAllowOrigin != p.Conf.MetricsAllowOrigin ||
		!reflect.DeepEqual(newConf.MetricsTrustedProxies, p.Conf.MetricsTrustedProxies) ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		closeAuthManager ||
		closeLogger

	closePPROF := newConf == nil ||
		newConf.PPROF != p.Conf.PPROF ||
		newConf.PPROFAddress != p.Conf.PPROFAddress ||
		newConf.PPROFEncryption != p.Conf.PPROFEncryption ||
		newConf.PPROFServerKey != p.Conf.PPROFServerKey ||
		newConf.PPROFServerCert != p.Conf.PPROFServerCert ||
		newConf.PPROFAllowOrigin != p.Conf.PPROFAllowOrigin ||
		!reflect.DeepEqual(newConf.PPROFTrustedProxies, p.Conf.PPROFTrustedProxies) ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		closeAuthManager ||
		closeLogger

	closeRecorderCleaner := newConf == nil ||
		atLeastOneRecordDeleteAfter(newConf.Paths) != atLeastOneRecordDeleteAfter(p.Conf.Paths) ||
		closeLogger
	if !closeRecorderCleaner && p.recordCleaner != nil && !reflect.DeepEqual(newConf.Paths, p.Conf.Paths) {
		p.recordCleaner.ReloadPathConfs(newConf.Paths)
	}

	closePlaybackServer := newConf == nil ||
		newConf.Playback != p.Conf.Playback ||
		newConf.PlaybackAddress != p.Conf.PlaybackAddress ||
		newConf.PlaybackEncryption != p.Conf.PlaybackEncryption ||
		newConf.PlaybackServerKey != p.Conf.PlaybackServerKey ||
		newConf.PlaybackServerCert != p.Conf.PlaybackServerCert ||
		newConf.PlaybackAllowOrigin != p.Conf.PlaybackAllowOrigin ||
		!reflect.DeepEqual(newConf.PlaybackTrustedProxies, p.Conf.PlaybackTrustedProxies) ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		closeAuthManager ||
		closeLogger
	if !closePlaybackServer && p.playbackServer != nil && !reflect.DeepEqual(newConf.Paths, p.Conf.Paths) {
		p.playbackServer.ReloadPathConfs(newConf.Paths)
	}

	closePathManager := newConf == nil ||
		newConf.LogLevel != p.Conf.LogLevel ||
		newConf.RTSPAddress != p.Conf.RTSPAddress ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		newConf.WriteTimeout != p.Conf.WriteTimeout ||
		newConf.WriteQueueSize != p.Conf.WriteQueueSize ||
		newConf.UDPMaxPayloadSize != p.Conf.UDPMaxPayloadSize ||
		newConf.RTSPEncryption != p.Conf.RTSPEncryption ||
		closeMetrics ||
		closeAuthManager ||
		closeLogger
	if !closePathManager && !reflect.DeepEqual(newConf.Paths, p.Conf.Paths) {
		p.PathManager.ReloadPathConfs(newConf.Paths)
	}

	closeRTSPServer := newConf == nil ||
		newConf.RTSP != p.Conf.RTSP ||
		newConf.RTSPEncryption != p.Conf.RTSPEncryption ||
		newConf.RTSPAddress != p.Conf.RTSPAddress ||
		!reflect.DeepEqual(newConf.RTSPAuthMethods, p.Conf.RTSPAuthMethods) ||
		newConf.RTSPUDPReadBufferSize != p.Conf.RTSPUDPReadBufferSize ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		newConf.WriteTimeout != p.Conf.WriteTimeout ||
		newConf.WriteQueueSize != p.Conf.WriteQueueSize ||
		newConf.RTPAddress != p.Conf.RTPAddress ||
		newConf.RTCPAddress != p.Conf.RTCPAddress ||
		newConf.MulticastIPRange != p.Conf.MulticastIPRange ||
		newConf.MulticastRTPPort != p.Conf.MulticastRTPPort ||
		newConf.MulticastRTCPPort != p.Conf.MulticastRTCPPort ||
		newConf.RTSPAddress != p.Conf.RTSPAddress ||
		!reflect.DeepEqual(newConf.RTSPTransports, p.Conf.RTSPTransports) ||
		newConf.RunOnConnect != p.Conf.RunOnConnect ||
		newConf.RunOnConnectRestart != p.Conf.RunOnConnectRestart ||
		newConf.RunOnDisconnect != p.Conf.RunOnDisconnect ||
		closeMetrics ||
		closePathManager ||
		closeLogger

	closeRTSPSServer := newConf == nil ||
		newConf.RTSP != p.Conf.RTSP ||
		newConf.RTSPEncryption != p.Conf.RTSPEncryption ||
		newConf.RTSPSAddress != p.Conf.RTSPSAddress ||
		!reflect.DeepEqual(newConf.RTSPAuthMethods, p.Conf.RTSPAuthMethods) ||
		newConf.RTSPUDPReadBufferSize != p.Conf.RTSPUDPReadBufferSize ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		newConf.WriteTimeout != p.Conf.WriteTimeout ||
		newConf.WriteQueueSize != p.Conf.WriteQueueSize ||
		newConf.RTSPServerCert != p.Conf.RTSPServerCert ||
		newConf.RTSPServerKey != p.Conf.RTSPServerKey ||
		newConf.RTSPAddress != p.Conf.RTSPAddress ||
		!reflect.DeepEqual(newConf.RTSPTransports, p.Conf.RTSPTransports) ||
		newConf.RunOnConnect != p.Conf.RunOnConnect ||
		newConf.RunOnConnectRestart != p.Conf.RunOnConnectRestart ||
		newConf.RunOnDisconnect != p.Conf.RunOnDisconnect ||
		closeMetrics ||
		closePathManager ||
		closeLogger

	closeRTMPServer := newConf == nil ||
		newConf.RTMP != p.Conf.RTMP ||
		newConf.RTMPEncryption != p.Conf.RTMPEncryption ||
		newConf.RTMPAddress != p.Conf.RTMPAddress ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		newConf.WriteTimeout != p.Conf.WriteTimeout ||
		newConf.RTSPAddress != p.Conf.RTSPAddress ||
		newConf.RunOnConnect != p.Conf.RunOnConnect ||
		newConf.RunOnConnectRestart != p.Conf.RunOnConnectRestart ||
		newConf.RunOnDisconnect != p.Conf.RunOnDisconnect ||
		closeMetrics ||
		closePathManager ||
		closeLogger

	closeRTMPSServer := newConf == nil ||
		newConf.RTMP != p.Conf.RTMP ||
		newConf.RTMPEncryption != p.Conf.RTMPEncryption ||
		newConf.RTMPSAddress != p.Conf.RTMPSAddress ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		newConf.WriteTimeout != p.Conf.WriteTimeout ||
		newConf.RTMPServerCert != p.Conf.RTMPServerCert ||
		newConf.RTMPServerKey != p.Conf.RTMPServerKey ||
		newConf.RTSPAddress != p.Conf.RTSPAddress ||
		newConf.RunOnConnect != p.Conf.RunOnConnect ||
		newConf.RunOnConnectRestart != p.Conf.RunOnConnectRestart ||
		newConf.RunOnDisconnect != p.Conf.RunOnDisconnect ||
		closeMetrics ||
		closePathManager ||
		closeLogger

	closeHLSServer := newConf == nil ||
		newConf.HLS != p.Conf.HLS ||
		newConf.HLSAddress != p.Conf.HLSAddress ||
		newConf.HLSEncryption != p.Conf.HLSEncryption ||
		newConf.HLSServerKey != p.Conf.HLSServerKey ||
		newConf.HLSServerCert != p.Conf.HLSServerCert ||
		newConf.HLSAllowOrigin != p.Conf.HLSAllowOrigin ||
		!reflect.DeepEqual(newConf.HLSTrustedProxies, p.Conf.HLSTrustedProxies) ||
		newConf.HLSAlwaysRemux != p.Conf.HLSAlwaysRemux ||
		newConf.HLSVariant != p.Conf.HLSVariant ||
		newConf.HLSSegmentCount != p.Conf.HLSSegmentCount ||
		newConf.HLSSegmentDuration != p.Conf.HLSSegmentDuration ||
		newConf.HLSPartDuration != p.Conf.HLSPartDuration ||
		newConf.HLSSegmentMaxSize != p.Conf.HLSSegmentMaxSize ||
		newConf.HLSDirectory != p.Conf.HLSDirectory ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		newConf.HLSMuxerCloseAfter != p.Conf.HLSMuxerCloseAfter ||
		closePathManager ||
		closeMetrics ||
		closeLogger

	closeWebRTCServer := newConf == nil ||
		newConf.WebRTC != p.Conf.WebRTC ||
		newConf.WebRTCAddress != p.Conf.WebRTCAddress ||
		newConf.WebRTCEncryption != p.Conf.WebRTCEncryption ||
		newConf.WebRTCServerKey != p.Conf.WebRTCServerKey ||
		newConf.WebRTCServerCert != p.Conf.WebRTCServerCert ||
		newConf.WebRTCAllowOrigin != p.Conf.WebRTCAllowOrigin ||
		!reflect.DeepEqual(newConf.WebRTCTrustedProxies, p.Conf.WebRTCTrustedProxies) ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		newConf.WebRTCLocalUDPAddress != p.Conf.WebRTCLocalUDPAddress ||
		newConf.WebRTCLocalTCPAddress != p.Conf.WebRTCLocalTCPAddress ||
		newConf.WebRTCIPsFromInterfaces != p.Conf.WebRTCIPsFromInterfaces ||
		!reflect.DeepEqual(newConf.WebRTCIPsFromInterfacesList, p.Conf.WebRTCIPsFromInterfacesList) ||
		!reflect.DeepEqual(newConf.WebRTCAdditionalHosts, p.Conf.WebRTCAdditionalHosts) ||
		!reflect.DeepEqual(newConf.WebRTCICEServers2, p.Conf.WebRTCICEServers2) ||
		newConf.WebRTCHandshakeTimeout != p.Conf.WebRTCHandshakeTimeout ||
		newConf.WebRTCSTUNGatherTimeout != p.Conf.WebRTCSTUNGatherTimeout ||
		newConf.WebRTCTrackGatherTimeout != p.Conf.WebRTCTrackGatherTimeout ||
		closeMetrics ||
		closePathManager ||
		closeLogger

	closeSRTServer := newConf == nil ||
		newConf.SRT != p.Conf.SRT ||
		newConf.SRTAddress != p.Conf.SRTAddress ||
		newConf.RTSPAddress != p.Conf.RTSPAddress ||
		newConf.ReadTimeout != p.Conf.ReadTimeout ||
		newConf.WriteTimeout != p.Conf.WriteTimeout ||
		newConf.UDPMaxPayloadSize != p.Conf.UDPMaxPayloadSize ||
		newConf.RunOnConnect != p.Conf.RunOnConnect ||
		newConf.RunOnConnectRestart != p.Conf.RunOnConnectRestart ||
		newConf.RunOnDisconnect != p.Conf.RunOnDisconnect ||
		closePathManager ||
		closeLogger

	// closeAPI := newConf == nil ||
	// 	newConf.API != p.Conf.API ||
	// 	newConf.APIAddress != p.Conf.APIAddress ||
	// 	newConf.APIEncryption != p.Conf.APIEncryption ||
	// 	newConf.APIServerKey != p.Conf.APIServerKey ||
	// 	newConf.APIServerCert != p.Conf.APIServerCert ||
	// 	newConf.APIAllowOrigin != p.Conf.APIAllowOrigin ||
	// 	!reflect.DeepEqual(newConf.APITrustedProxies, p.Conf.APITrustedProxies) ||
	// 	newConf.ReadTimeout != p.Conf.ReadTimeout ||
	// 	closeAuthManager ||
	// 	closePathManager ||
	// 	closeRTSPServer ||
	// 	closeRTSPSServer ||
	// 	closeRTMPServer ||
	// 	closeHLSServer ||
	// 	closeWebRTCServer ||
	// 	closeSRTServer ||
	// 	closeLogger

	if newConf == nil && p.ConfWatcher != nil {
		p.ConfWatcher.Close()
		p.ConfWatcher = nil
	}

	// if p.api != nil {
	// 	if closeAPI {
	// 		p.api.Close()
	// 		p.api = nil
	// 	} else if !calledByAPI { // avoid a loop
	// 		p.api.ReloadConf(newConf)
	// 	}
	// }

	if closeSRTServer && p.SrtServer != nil {
		p.SrtServer.Close()
		p.SrtServer = nil
	}

	if closeWebRTCServer && p.WebRTCServer != nil {
		p.WebRTCServer.Close()
		p.WebRTCServer = nil
	}

	if closeHLSServer && p.HlsServer != nil {
		p.HlsServer.Close()
		p.HlsServer = nil
	}

	if closeRTMPSServer && p.RtmpsServer != nil {
		p.RtmpsServer.Close()
		p.RtmpsServer = nil
	}

	if closeRTMPServer && p.RtmpServer != nil {
		p.RtmpServer.Close()
		p.RtmpServer = nil
	}

	if closeRTSPSServer && p.RtspsServer != nil {
		p.RtspsServer.Close()
		p.RtspsServer = nil
	}

	if closeRTSPServer && p.RtspServer != nil {
		p.RtspServer.Close()
		p.RtspServer = nil
	}

	if closePathManager && p.PathManager != nil {
		p.PathManager.close()
		p.PathManager = nil
	}

	if closePlaybackServer && p.playbackServer != nil {
		p.playbackServer.Close()
		p.playbackServer = nil
	}

	if closeRecorderCleaner && p.recordCleaner != nil {
		p.recordCleaner.Close()
		p.recordCleaner = nil
	}

	if closePPROF && p.Pprof != nil {
		p.Pprof.Close()
		p.Pprof = nil
	}

	if closeMetrics && p.Metrics != nil {
		p.Metrics.Close()
		p.Metrics = nil
	}

	if closeAuthManager && p.AuthManager != nil {
		p.AuthManager = nil
	}

	if newConf == nil && p.ExternalCmdPool != nil {
		p.Log(logger.Info, "waiting for running hooks")
		p.ExternalCmdPool.Close()
	}

	if closeLogger && p.Logger != nil {
		p.Logger.Close()
		p.Logger = nil
	}
}

func (p *Core) reloadConf(newConf *conf.Conf, calledByAPI bool) error {
	p.closeResources(newConf, calledByAPI)
	p.Conf = newConf
	return p.CreateResources(false)
}

// APIConfigSet is called by api.
func (p *Core) APIConfigSet(conf *conf.Conf) {
	select {
	case p.ChAPIConfigSet <- conf:
	case <-p.ctx.Done():
	}
}
