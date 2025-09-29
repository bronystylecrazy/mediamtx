package mtx

// Generated structs based on api/openapi.yaml models
// These provide type-safe alternatives to JSON string manipulation

import (
	"encoding/json"
	"fmt"
	"github.com/bluenviron/mediamtx/pkg/conf"
	"time"
)

// APIError represents an error response from the API
type APIError struct {
	Error string `json:"error"`
}

// AuthInternalUser represents an internal authentication user
type AuthInternalUser struct {
	User        string                       `json:"user,omitempty"`
	Pass        string                       `json:"pass,omitempty"`
	IPs         []string                     `json:"ips,omitempty"`
	Permissions []AuthInternalUserPermission `json:"permissions,omitempty"`
}

// AuthInternalUserPermission represents user permissions
type AuthInternalUserPermission struct {
	Action string `json:"action,omitempty"`
	Path   string `json:"PathHandler,omitempty"`
}

// GlobalConf represents the global configuration
type GlobalConf struct {
	// General
	LogLevel            string   `json:"logLevel,omitempty"`
	LogDestinations     []string `json:"logDestinations,omitempty"`
	LogFile             string   `json:"logFile,omitempty"`
	SysLogPrefix        string   `json:"sysLogPrefix,omitempty"`
	ReadTimeout         string   `json:"readTimeout,omitempty"`
	WriteTimeout        string   `json:"writeTimeout,omitempty"`
	WriteQueueSize      int64    `json:"writeQueueSize,omitempty"`
	UDPMaxPayloadSize   int64    `json:"udpMaxPayloadSize,omitempty"`
	RunOnConnect        string   `json:"runOnConnect,omitempty"`
	RunOnConnectRestart bool     `json:"runOnConnectRestart,omitempty"`
	RunOnDisconnect     string   `json:"runOnDisconnect,omitempty"`

	// Authentication
	AuthMethod             string                       `json:"authMethod,omitempty"`
	AuthInternalUsers      []AuthInternalUser           `json:"authInternalUsers,omitempty"`
	AuthHTTPAddress        string                       `json:"authHTTPAddress,omitempty"`
	AuthHTTPExclude        []AuthInternalUserPermission `json:"authHTTPExclude,omitempty"`
	AuthJWTJWKS            string                       `json:"authJWTJWKS,omitempty"`
	AuthJWTJWKSFingerprint string                       `json:"authJWTJWKSFingerprint,omitempty"`
	AuthJWTClaimKey        string                       `json:"authJWTClaimKey,omitempty"`
	AuthJWTExclude         []AuthInternalUserPermission `json:"authJWTExclude,omitempty"`
}

// PathConf represents a PathHandler configuration (matches OpenAPI PathConf)
type PathConf struct {
	Name string `json:"name,omitempty"`

	// General
	Source                     string `json:"source,omitempty"`
	SourceFingerprint          string `json:"sourceFingerprint,omitempty"`
	SourceOnDemand             bool   `json:"sourceOnDemand,omitempty"`
	SourceOnDemandStartTimeout string `json:"sourceOnDemandStartTimeout,omitempty"`
	SourceOnDemandCloseAfter   string `json:"sourceOnDemandCloseAfter,omitempty"`
	RunOnDemand                string `json:"runOnDemand,omitempty"`
	MaxReaders                 int64  `json:"maxReaders,omitempty"`
	SRTReadPassphrase          string `json:"srtReadPassphrase,omitempty"`
	Fallback                   string `json:"fallback,omitempty"`
	UseAbsoluteTimestamp       bool   `json:"useAbsoluteTimestamp,omitempty"`

	// Record
	Record                bool   `json:"record,omitempty"`
	RecordPath            string `json:"recordPath,omitempty"`
	RecordFormat          string `json:"recordFormat,omitempty"`
	RecordPartDuration    string `json:"recordPartDuration,omitempty"`
	RecordMaxPartSize     string `json:"recordMaxPartSize,omitempty"`
	RecordSegmentDuration string `json:"recordSegmentDuration,omitempty"`
	RecordDeleteAfter     string `json:"recordDeleteAfter,omitempty"`

	// Publisher source
	OverridePublisher    bool   `json:"overridePublisher,omitempty"`
	SRTPublishPassphrase string `json:"srtPublishPassphrase,omitempty"`

	// RTSP source
	RTSPTransport         string `json:"rtspTransport,omitempty"`
	RTSPAnyPort           bool   `json:"rtspAnyPort,omitempty"`
	RTSPRangeType         string `json:"rtspRangeType,omitempty"`
	RTSPRangeStart        string `json:"rtspRangeStart,omitempty"`
	RTSPUDPReadBufferSize int64  `json:"rtspUDPReadBufferSize,omitempty"`

	// MPEG-TS source
	MPEGTSUDPReadBufferSize int64 `json:"mpegtsUDPReadBufferSize,omitempty"`

	// RTP source
	RTPSDP               string `json:"rtpSDP,omitempty"`
	RTPUDPReadBufferSize int64  `json:"rtpUDPReadBufferSize,omitempty"`

	// Redirect source
	SourceRedirect string `json:"sourceRedirect,omitempty"`

	// Raspberry Pi Camera source
	RPiCameraCamID        int64     `json:"rpiCameraCamID,omitempty"`
	RPiCameraSecondary    bool      `json:"rpiCameraSecondary,omitempty"`
	RPiCameraWidth        int64     `json:"rpiCameraWidth,omitempty"`
	RPiCameraHeight       int64     `json:"rpiCameraHeight,omitempty"`
	RPiCameraHFlip        bool      `json:"rpiCameraHFlip,omitempty"`
	RPiCameraVFlip        bool      `json:"rpiCameraVFlip,omitempty"`
	RPiCameraBrightness   float64   `json:"rpiCameraBrightness,omitempty"`
	RPiCameraContrast     float64   `json:"rpiCameraContrast,omitempty"`
	RPiCameraSaturation   float64   `json:"rpiCameraSaturation,omitempty"`
	RPiCameraSharpness    float64   `json:"rpiCameraSharpness,omitempty"`
	RPiCameraExposure     string    `json:"rpiCameraExposure,omitempty"`
	RPiCameraAWB          string    `json:"rpiCameraAWB,omitempty"`
	RPiCameraAWBGains     []float64 `json:"rpiCameraAWBGains,omitempty"`
	RPiCameraDenoise      string    `json:"rpiCameraDenoise,omitempty"`
	RPiCameraShutter      int64     `json:"rpiCameraShutter,omitempty"`
	RPiCameraMetering     string    `json:"rpiCameraMetering,omitempty"`
	RPiCameraGain         float64   `json:"rpiCameraGain,omitempty"`
	RPiCameraEV           float64   `json:"rpiCameraEV,omitempty"`
	RPiCameraROI          string    `json:"rpiCameraROI,omitempty"`
	RPiCameraHDR          bool      `json:"rpiCameraHDR,omitempty"`
	RPiCameraTuningFile   string    `json:"rpiCameraTuningFile,omitempty"`
	RPiCameraMode         string    `json:"rpiCameraMode,omitempty"`
	RPiCameraFPS          float64   `json:"rpiCameraFPS,omitempty"`
	RPiCameraAfMode       string    `json:"rpiCameraAfMode,omitempty"`
	RPiCameraAfRange      string    `json:"rpiCameraAfRange,omitempty"`
	RPiCameraAfSpeed      string    `json:"rpiCameraAfSpeed,omitempty"`
	RPiCameraLensPosition float64   `json:"rpiCameraLensPosition,omitempty"`
	RPiCameraAfWindow     string    `json:"rpiCameraAfWindow,omitempty"`
}

// PathConfList represents a paginated list of PathHandler configurations
type PathConfList struct {
	PageCount int64      `json:"pageCount"`
	ItemCount int64      `json:"itemCount"`
	Items     []PathConf `json:"items"`
}

// PathSource represents a PathHandler source
type PathSource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// PathSourceType represents the type of PathHandler source
type PathSourceType string

const (
	PathSourceTypeHLSSource       PathSourceType = "hlsSource"
	PathSourceTypeRedirect        PathSourceType = "redirect"
	PathSourceTypeRPiCameraSource PathSourceType = "rpiCameraSource"
	PathSourceTypeRTMPConn        PathSourceType = "rtmpConn"
	PathSourceTypeRTMPSource      PathSourceType = "rtmpSource"
	PathSourceTypeRTSPSession     PathSourceType = "rtspSession"
	PathSourceTypeRTSPSource      PathSourceType = "rtspSource"
	PathSourceTypeRTSPSSession    PathSourceType = "rtspsSession"
	PathSourceTypeSRTConn         PathSourceType = "srtConn"
	PathSourceTypeSRTSource       PathSourceType = "srtSource"
	PathSourceTypeMPEGTSSource    PathSourceType = "mpegtsSource"
	PathSourceTypeRTPSource       PathSourceType = "rtpSource"
	PathSourceTypeWebRTCSession   PathSourceType = "webRTCSession"
	PathSourceTypeWebRTCSource    PathSourceType = "webRTCSource"
)

// PathReader represents a PathHandler reader
type PathReader struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Path represents runtime PathHandler information
type Path struct {
	Name          string       `json:"name"`
	ConfName      string       `json:"confName"`
	Source        *PathSource  `json:"source"`
	Ready         bool         `json:"ready"`
	ReadyTime     *time.Time   `json:"readyTime"`
	Tracks        []string     `json:"tracks"`
	BytesReceived int64        `json:"bytesReceived"`
	BytesSent     int64        `json:"bytesSent"`
	Readers       []PathReader `json:"readers"`
}

// PathList represents a paginated list of runtime paths
type PathList struct {
	PageCount int64  `json:"pageCount"`
	ItemCount int64  `json:"itemCount"`
	Items     []Path `json:"items"`
}

// HLSMuxer represents an HLS muxer
type HLSMuxer struct {
	Path        string    `json:"PathHandler"`
	Created     time.Time `json:"created"`
	LastRequest time.Time `json:"lastRequest"`
	BytesSent   int64     `json:"bytesSent"`
}

// HLSMuxerList represents a list of HLS muxers
type HLSMuxerList struct {
	PageCount int64      `json:"pageCount"`
	ItemCount int64      `json:"itemCount"`
	Items     []HLSMuxer `json:"items"`
}

// Recording represents a recording
type Recording struct {
	Name     string             `json:"name"`
	Segments []RecordingSegment `json:"segments"`
}

// RecordingList represents a list of recordings
type RecordingList struct {
	PageCount int64       `json:"pageCount"`
	ItemCount int64       `json:"itemCount"`
	Items     []Recording `json:"items"`
}

// RecordingSegment represents a recording segment
type RecordingSegment struct {
	Start time.Time `json:"start"`
}

// RTMPConn represents an RTMP connection
type RTMPConn struct {
	ID            string    `json:"id"`
	Created       time.Time `json:"created"`
	RemoteAddr    string    `json:"remoteAddr"`
	State         string    `json:"state"`
	Path          string    `json:"PathHandler"`
	Query         string    `json:"query"`
	BytesReceived int64     `json:"bytesReceived"`
	BytesSent     int64     `json:"bytesSent"`
}

// RTMPConnList represents a list of RTMP connections
type RTMPConnList struct {
	PageCount int64      `json:"pageCount"`
	ItemCount int64      `json:"itemCount"`
	Items     []RTMPConn `json:"items"`
}

// RTSPConn represents an RTSP connection
type RTSPConn struct {
	ID            string    `json:"id"`
	Created       time.Time `json:"created"`
	RemoteAddr    string    `json:"remoteAddr"`
	BytesReceived int64     `json:"bytesReceived"`
	BytesSent     int64     `json:"bytesSent"`
	Session       *string   `json:"session"`
	Tunnel        string    `json:"tunnel"`
}

// RTSPConnList represents a list of RTSP connections
type RTSPConnList struct {
	PageCount int64      `json:"pageCount"`
	ItemCount int64      `json:"itemCount"`
	Items     []RTSPConn `json:"items"`
}

// RTSPSession represents an RTSP session
type RTSPSession struct {
	ID                  string    `json:"id"`
	Created             time.Time `json:"created"`
	RemoteAddr          string    `json:"remoteAddr"`
	State               string    `json:"state"`
	Path                string    `json:"PathHandler"`
	Query               string    `json:"query"`
	Transport           *string   `json:"transport"`
	Profile             *string   `json:"profile"`
	BytesReceived       int64     `json:"bytesReceived"`
	BytesSent           int64     `json:"bytesSent"`
	RTPPacketsReceived  int64     `json:"rtpPacketsReceived"`
	RTPPacketsSent      int64     `json:"rtpPacketsSent"`
	RTPPacketsLost      int64     `json:"rtpPacketsLost"`
	RTPPacketsInError   int64     `json:"rtpPacketsInError"`
	RTPPacketsJitter    float64   `json:"rtpPacketsJitter"`
	RTCPPacketsReceived int64     `json:"rtcpPacketsReceived"`
	RTCPPacketsSent     int64     `json:"rtcpPacketsSent"`
	RTCPPacketsInError  int64     `json:"rtcpPacketsInError"`
}

// RTSPSessionList represents a list of RTSP sessions
type RTSPSessionList struct {
	PageCount int64         `json:"pageCount"`
	ItemCount int64         `json:"itemCount"`
	Items     []RTSPSession `json:"items"`
}

// SRTConn represents an SRT connection
type SRTConn struct {
	ID                            string    `json:"id"`
	Created                       time.Time `json:"created"`
	RemoteAddr                    string    `json:"remoteAddr"`
	State                         string    `json:"state"`
	Path                          string    `json:"PathHandler"`
	Query                         string    `json:"query"`
	PacketsSent                   int64     `json:"packetsSent"`
	PacketsReceived               int64     `json:"packetsReceived"`
	PacketsSentUnique             int64     `json:"packetsSentUnique"`
	PacketsReceivedUnique         int64     `json:"packetsReceivedUnique"`
	PacketsSendLoss               int64     `json:"packetsSendLoss"`
	PacketsReceivedLoss           int64     `json:"packetsReceivedLoss"`
	PacketsRetrans                int64     `json:"packetsRetrans"`
	PacketsReceivedRetrans        int64     `json:"packetsReceivedRetrans"`
	PacketsSentACK                int64     `json:"packetsSentACK"`
	PacketsReceivedACK            int64     `json:"packetsReceivedACK"`
	PacketsSentNAK                int64     `json:"packetsSentNAK"`
	PacketsReceivedNAK            int64     `json:"packetsReceivedNAK"`
	PacketsSentKM                 int64     `json:"packetsSentKM"`
	PacketsReceivedKM             int64     `json:"packetsReceivedKM"`
	UsSndDuration                 int64     `json:"usSndDuration"`
	PacketsReceivedBelated        int64     `json:"packetsReceivedBelated"`
	PacketsSendDrop               int64     `json:"packetsSendDrop"`
	PacketsReceivedDrop           int64     `json:"packetsReceivedDrop"`
	PacketsReceivedUndecrypt      int64     `json:"packetsReceivedUndecrypt"`
	BytesReceived                 int64     `json:"bytesReceived"`
	BytesSent                     int64     `json:"bytesSent"`
	BytesSentUnique               int64     `json:"bytesSentUnique"`
	BytesReceivedUnique           int64     `json:"bytesReceivedUnique"`
	BytesReceivedLoss             int64     `json:"bytesReceivedLoss"`
	BytesRetrans                  int64     `json:"bytesRetrans"`
	BytesReceivedRetrans          int64     `json:"bytesReceivedRetrans"`
	BytesReceivedBelated          int64     `json:"bytesReceivedBelated"`
	BytesSendDrop                 int64     `json:"bytesSendDrop"`
	BytesReceivedDrop             int64     `json:"bytesReceivedDrop"`
	BytesReceivedUndecrypt        int64     `json:"bytesReceivedUndecrypt"`
	UsPacketsSendPeriod           float64   `json:"usPacketsSendPeriod"`
	PacketsFlowWindow             int64     `json:"packetsFlowWindow"`
	PacketsFlightSize             int64     `json:"packetsFlightSize"`
	MsRTT                         float64   `json:"msRTT"`
	MbpsSendRate                  float64   `json:"mbpsSendRate"`
	MbpsReceiveRate               float64   `json:"mbpsReceiveRate"`
	MbpsLinkCapacity              float64   `json:"mbpsLinkCapacity"`
	BytesAvailSendBuf             int64     `json:"bytesAvailSendBuf"`
	BytesAvailReceiveBuf          int64     `json:"bytesAvailReceiveBuf"`
	MbpsMaxBW                     float64   `json:"mbpsMaxBW"`
	ByteMSS                       int64     `json:"byteMSS"`
	PacketsSendBuf                int64     `json:"packetsSendBuf"`
	BytesSendBuf                  int64     `json:"bytesSendBuf"`
	MsSendBuf                     int64     `json:"msSendBuf"`
	MsSendTsbPdDelay              int64     `json:"msSendTsbPdDelay"`
	PacketsReceiveBuf             int64     `json:"packetsReceiveBuf"`
	BytesReceiveBuf               int64     `json:"bytesReceiveBuf"`
	MsReceiveBuf                  int64     `json:"msReceiveBuf"`
	MsReceiveTsbPdDelay           int64     `json:"msReceiveTsbPdDelay"`
	PacketsReorderTolerance       int64     `json:"packetsReorderTolerance"`
	PacketsReceivedAvgBelatedTime int64     `json:"packetsReceivedAvgBelatedTime"`
	PacketsSendLossRate           float64   `json:"packetsSendLossRate"`
	PacketsReceivedLossRate       float64   `json:"packetsReceivedLossRate"`
}

// SRTConnList represents a list of SRT connections
type SRTConnList struct {
	PageCount int64     `json:"pageCount"`
	ItemCount int64     `json:"itemCount"`
	Items     []SRTConn `json:"items"`
}

// WebRTCSession represents a WebRTC session
type WebRTCSession struct {
	ID                        string    `json:"id"`
	Created                   time.Time `json:"created"`
	RemoteAddr                string    `json:"remoteAddr"`
	PeerConnectionEstablished bool      `json:"peerConnectionEstablished"`
	LocalCandidate            string    `json:"localCandidate"`
	RemoteCandidate           string    `json:"remoteCandidate"`
	State                     string    `json:"state"`
	Path                      string    `json:"PathHandler"`
	Query                     string    `json:"query"`
	BytesReceived             int64     `json:"bytesReceived"`
	BytesSent                 int64     `json:"bytesSent"`
	RTPPacketsReceived        int64     `json:"rtpPacketsReceived"`
	RTPPacketsSent            int64     `json:"rtpPacketsSent"`
	RTPPacketsLost            int64     `json:"rtpPacketsLost"`
	RTPPacketsJitter          float64   `json:"rtpPacketsJitter"`
	RTCPPacketsReceived       int64     `json:"rtcpPacketsReceived"`
	RTCPPacketsSent           int64     `json:"rtcpPacketsSent"`
}

// WebRTCSessionList represents a list of WebRTC sessions
type WebRTCSessionList struct {
	PageCount int64           `json:"pageCount"`
	ItemCount int64           `json:"itemCount"`
	Items     []WebRTCSession `json:"items"`
}

// PathConfigBuilder provides a fluent API for building PathHandler configurations
type PathConfigBuilder struct {
	config PathConf
}

// NewPathConfigBuilder creates a new PathHandler configuration builder
func NewPathConfigBuilder() *PathConfigBuilder {
	return &PathConfigBuilder{
		config: PathConf{},
	}
}

// Convenience factory methods for common PathHandler types

// NewSimplePathConfig creates a builder for a simple PathHandler
func NewSimplePathConfig(name, source string, enableRecording bool) *PathConfigBuilder {
	builder := NewPathConfigBuilder().SetName(name).SetSource(source)
	if enableRecording {
		builder = builder.SetupRecording(name)
	}
	return builder
}

// NewRTSPPathConfig creates a builder for an RTSP PathHandler
func NewRTSPPathConfig(name, rtspURL string, enableRecording bool) *PathConfigBuilder {
	builder := NewPathConfigBuilder().SetName(name).SetSource(rtspURL).SetupRTSP()
	if enableRecording {
		builder = builder.SetupRecording(name)
	}
	return builder
}

// NewPublisherPathConfig creates a builder for a publisher PathHandler
func NewPublisherPathConfig(name string, enableRecording bool) *PathConfigBuilder {
	builder := NewPathConfigBuilder().SetName(name).SetSource("publisher")
	if enableRecording {
		builder = builder.SetupRecording(name)
	}
	return builder
}

// NewOnDemandPathConfig creates a builder for an on-demand PathHandler
func NewOnDemandPathConfig(name, source, command string) *PathConfigBuilder {
	return NewPathConfigBuilder().
		SetName(name).
		SetSource(source).
		SetupOnDemand(command)
}

// SetName sets the PathHandler name
func (b *PathConfigBuilder) SetName(name string) *PathConfigBuilder {
	b.config.Name = name
	return b
}

// SetSource sets the PathHandler source
func (b *PathConfigBuilder) SetSource(source string) *PathConfigBuilder {
	b.config.Source = source
	return b
}

// SetRecording enables or disables recording
func (b *PathConfigBuilder) SetRecording(enabled bool) *PathConfigBuilder {
	b.config.Record = enabled
	return b
}

// SetRecordPath sets the recording PathHandler
func (b *PathConfigBuilder) SetRecordPath(path string) *PathConfigBuilder {
	b.config.RecordPath = path
	return b
}

// SetRecordFormat sets the recording format
func (b *PathConfigBuilder) SetRecordFormat(format string) *PathConfigBuilder {
	b.config.RecordFormat = format
	return b
}

// SetMaxReaders sets the maximum number of readers
func (b *PathConfigBuilder) SetMaxReaders(max int64) *PathConfigBuilder {
	b.config.MaxReaders = max
	return b
}

// SetSourceOnDemand enables or disables on-demand source activation
func (b *PathConfigBuilder) SetSourceOnDemand(enabled bool) *PathConfigBuilder {
	b.config.SourceOnDemand = enabled
	return b
}

// SetRunOnDemand sets the command to run on demand
func (b *PathConfigBuilder) SetRunOnDemand(command string) *PathConfigBuilder {
	b.config.RunOnDemand = command
	return b
}

// SetRTSPTransport sets the RTSP transport mode
func (b *PathConfigBuilder) SetRTSPTransport(transport string) *PathConfigBuilder {
	b.config.RTSPTransport = transport
	return b
}

// SetRTSPAnyPort enables or disables RTSP any port mode
func (b *PathConfigBuilder) SetRTSPAnyPort(enabled bool) *PathConfigBuilder {
	b.config.RTSPAnyPort = enabled
	return b
}

// Build returns the built PathHandler configuration
// Convenience preset methods for common configurations

// SetupRecording configures recording with default settings
func (b *PathConfigBuilder) SetupRecording(pathName string) *PathConfigBuilder {
	return b.SetRecording(true).
		SetRecordPath(fmt.Sprintf("/recordings/%s", pathName)).
		SetRecordFormat("fmp4")
}

// SetupRTSP configures RTSP-specific settings with defaults
func (b *PathConfigBuilder) SetupRTSP() *PathConfigBuilder {
	return b.SetRTSPTransport("automatic").SetSourceOnDemand(false)
}

// SetupOnDemand configures on-demand activation
func (b *PathConfigBuilder) SetupOnDemand(command string) *PathConfigBuilder {
	return b.SetSourceOnDemand(true).SetRunOnDemand(command)
}

func (b *PathConfigBuilder) Build() PathConf {
	return b.config
}

// ToOptionalPath converts PathConf to conf.OptionalPath using JSON marshaling
func (pc *PathConf) ToOptionalPath() (*conf.OptionalPath, error) {
	jsonData, err := json.Marshal(pc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PathConf: %v", err)
	}

	var optPath conf.OptionalPath
	if err := json.Unmarshal(jsonData, &optPath); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to OptionalPath: %v", err)
	}

	return &optPath, nil
}

// ToConfPath converts PathConf to *conf.Path using JSON marshaling
func (pc *PathConf) ToConfPath() (*conf.Path, error) {
	jsonData, err := json.Marshal(pc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PathConf: %v", err)
	}

	var confPath conf.Path
	if err := json.Unmarshal(jsonData, &confPath); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to conf.Path: %v", err)
	}

	return &confPath, nil
}

// ConvertPathConfSliceToConfPaths converts []PathConf to []*conf.Path
func ConvertPathConfSliceToConfPaths(pathConfs []PathConf) ([]*conf.Path, error) {
	confPaths := make([]*conf.Path, 0, len(pathConfs))

	for _, pathConf := range pathConfs {
		confPath, err := pathConf.ToConfPath()
		if err != nil {
			continue // Skip invalid paths
		}
		confPaths = append(confPaths, confPath)
	}

	return confPaths, nil
}

// PathConfigFromJSON creates a PathConf from JSON string
func PathConfigFromJSON(jsonStr string) (*PathConf, error) {
	var config PathConf
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return &config, nil
}

// ToJSON converts PathConf to JSON string
func (pc *PathConf) ToJSON() (string, error) {
	jsonData, err := json.Marshal(pc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %v", err)
	}
	return string(jsonData), nil
}
