package mtx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bluenviron/mediamtx/pkg/auth"
	conf2 "github.com/bluenviron/mediamtx/pkg/conf"
	defs2 "github.com/bluenviron/mediamtx/pkg/defs"
	"net"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MediaMTXAPIInterface defines the contract for MediaMTX API operations
type MediaMTXAPIInterface interface {
	// Configuration Management
	GetGlobalConfig() (*conf2.Conf, error)
	UpdateGlobalConfig(newConf *conf2.Conf) error
	PatchGlobalConfig(optionalGlobal *conf2.OptionalGlobal) error
	GetPathDefaults() *conf2.Path
	UpdatePathDefaults(defaults *conf2.OptionalPath) error
	PatchPathDefaults(optionalPath *conf2.OptionalPath) error

	// Path Configuration Management
	ListPathConfigs(pagination *PaginationParams) (*MediaMTXPathConfList, error)
	GetPathConfig(name string) (*conf2.Path, error)
	AddPathConfig(name string, pathConf *conf2.OptionalPath) error
	UpdatePathConfig(name string, pathConf *conf2.OptionalPath) error
	ReplacePathConfig(name string, pathConf *conf2.OptionalPath) error
	DeletePathConfig(name string) error

	// Runtime Path Information
	ListActivePaths(pagination *PaginationParams) (*defs2.APIPathList, error)
	GetActivePath(name string) (*defs2.APIPath, error)

	// RTSP Server Management
	GetRTSPConnections(pagination *PaginationParams) (*defs2.APIRTSPConnsList, error)
	GetRTSPConnection(id string) (*defs2.APIRTSPConn, error)
	GetRTSPSessions(pagination *PaginationParams) (*defs2.APIRTSPSessionList, error)
	GetRTSPSession(id string) (*defs2.APIRTSPSession, error)
	KickRTSPSession(id string) error

	// RTMP Server Management
	GetRTMPConnections(pagination *PaginationParams) (*defs2.APIRTMPConnList, error)
	GetRTMPConnection(id string) (*defs2.APIRTMPConn, error)
	KickRTMPConnection(id string) error

	// RTMPS Server Management
	GetRTMPSConnections(pagination *PaginationParams) (*defs2.APIRTMPConnList, error)
	GetRTMPSConnection(id string) (*defs2.APIRTMPConn, error)
	KickRTMPSConnection(id string) error

	// HLS Server Management
	GetHLSMuxers(pagination *PaginationParams) (*defs2.APIHLSMuxerList, error)
	GetHLSMuxer(name string) (*defs2.APIHLSMuxer, error)

	// WebRTC Server Management
	GetWebRTCSessions(pagination *PaginationParams) (*defs2.APIWebRTCSessionList, error)
	GetWebRTCSession(id string) (*defs2.APIWebRTCSession, error)
	KickWebRTCSession(id string) error

	// SRT Server Management
	GetSRTConnections(pagination *PaginationParams) (*defs2.APISRTConnList, error)
	GetSRTConnection(id string) (*defs2.APISRTConn, error)
	KickSRTConnection(id string) error

	// RTSPS Server Management
	GetRTSPSConnections(pagination *PaginationParams) (*defs2.APIRTSPConnsList, error)
	GetRTSPSConnection(id string) (*defs2.APIRTSPConn, error)
	GetRTSPSSessions(pagination *PaginationParams) (*defs2.APIRTSPSessionList, error)
	GetRTSPSSession(id string) (*defs2.APIRTSPSession, error)
	KickRTSPSSession(id string) error

	// Recording Management
	GetRecordings(query *RecordingQuery, pagination *PaginationParams) (*defs2.APIRecordingList, error)
	GetRecording(pathName string) (*defs2.APIRecording, error)
	DeleteRecordingSegment(pathName string, segmentStart time.Time) error
	GetRecordingsByPath(pathName string, pagination *PaginationParams) (*defs2.APIRecordingList, error)
	GetRecordingsByTimeRange(startTime, endTime time.Time, pagination *PaginationParams) (*defs2.APIRecordingList, error)
	GetRecordingInfo(pathName string) (*RecordingInfo, error)

	// Recording Operations
	StartRecording(pathName string) error
	StopRecording(pathName string) error
	IsRecording(pathName string) (bool, error)
	SetRecordingPath(pathName, recordingPath string) error
	GetRecordingPath(pathName string) (string, error)

	// Authentication Management
	Authenticate(req *auth.Request) *auth.Error
	RefreshJWTJWKS()
	CreateAuthRequest(user, pass, query, ip string) (*auth.Request, error)
	ValidateAPIAccess(user, pass, ip string) error
}

// MediaMTXAPI provides programmatic access to MediaMTX functionality without HTTP/gin dependencies
type MediaMTXAPI struct {
	core  *Core
	mutex sync.RWMutex
}

// Ensure MediaMTXAPI implements MediaMTXAPIInterface
var _ MediaMTXAPIInterface = (*MediaMTXAPI)(nil)

// NewMediaMTXAPI creates a new MediaMTX API instance
func NewMediaMTXAPI(core *Core) *MediaMTXAPI {
	return &MediaMTXAPI{
		core: core,
	}
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	ItemsPerPage int
	Page         int
}

// DefaultPagination returns default pagination settings
func DefaultPagination() *PaginationParams {
	return &PaginationParams{
		ItemsPerPage: 100,
		Page:         0,
	}
}

// PaginationResult contains pagination metadata
type PaginationResult struct {
	ItemCount int
	PageCount int
}

// APIResult represents a generic API result with optional pagination
type APIResult struct {
	Data       interface{}
	Pagination *PaginationResult
	Error      error
}

// MediaMTXPathConfList represents a list of PathHandler configurations for MediaMTX API
type MediaMTXPathConfList struct {
	ItemCount int                      `json:"itemCount"`
	PageCount int                      `json:"pageCount"`
	Items     []map[string]interface{} `json:"items"`
}

// =============================================================================
// CONFIGURATION MANAGEMENT
// =============================================================================

// GetGlobalConfig returns the current global configuration
func (api *MediaMTXAPI) GetGlobalConfig() (*conf2.Conf, error) {
	api.mutex.RLock()
	defer api.mutex.RUnlock()

	if api.core.Conf == nil {
		return nil, fmt.Errorf("configuration not available")
	}
	return api.core.Conf.Clone(), nil
}

// UpdateGlobalConfig updates the global configuration
func (api *MediaMTXAPI) UpdateGlobalConfig(newConf *conf2.Conf) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	// Validate configuration
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	// Apply configuration
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)

	return nil
}

// PatchGlobalConfig patches the global configuration (equivalent to PATCH /config/global/patch)
func (api *MediaMTXAPI) PatchGlobalConfig(optionalGlobal *conf2.OptionalGlobal) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	newConf := api.core.Conf.Clone()

	// Apply the patch
	newConf.PatchGlobal(optionalGlobal)

	// Validate the new configuration
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	// Apply configuration
	api.core.Conf = newConf

	// Use goroutine like the original API since config reload can cause shutdown
	go api.core.APIConfigSet(newConf)

	return nil
}

// GetPathDefaults returns the default PathHandler configuration
func (api *MediaMTXAPI) GetPathDefaults() *conf2.Path {
	return &conf2.Path{}
}

// UpdatePathDefaults updates the default PathHandler configuration
func (api *MediaMTXAPI) UpdatePathDefaults(defaults *conf2.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	newConf := api.core.Conf.Clone()

	// Apply defaults to all paths or handle as needed
	// This might need specific implementation based on requirements

	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)

	return nil
}

// PatchPathDefaults patches the default PathHandler configuration (equivalent to PATCH /config/pathdefaults/patch)
func (api *MediaMTXAPI) PatchPathDefaults(optionalPath *conf2.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	newConf := api.core.Conf.Clone()

	// Apply the patch to PathHandler defaults
	newConf.PatchPathDefaults(optionalPath)

	// Validate the new configuration
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	// Apply configuration
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)

	return nil
}

// =============================================================================
// PATH CONFIGURATION MANAGEMENT
// =============================================================================

// ListPathConfigs returns a list of all configured paths with pagination
func (api *MediaMTXAPI) ListPathConfigs(pagination *PaginationParams) (*MediaMTXPathConfList, error) {
	api.mutex.RLock()
	conf := api.core.Conf
	api.mutex.RUnlock()

	if conf == nil {
		return nil, fmt.Errorf("configuration not available")
	}

	// Create sorted list of paths
	sortedNames := api.sortedPathKeys(conf.Paths)
	data := &MediaMTXPathConfList{
		Items: []map[string]interface{}{},
	}

	for _, name := range sortedNames {
		// Convert PathHandler to map for JSON serialization
		pathMap := map[string]interface{}{
			"name":                       name,
			"source":                     conf.Paths[name].Source,
			"record":                     conf.Paths[name].Record,
			"recordPath":                 conf.Paths[name].RecordPath,
			"recordFormat":               conf.Paths[name].RecordFormat,
			"maxReaders":                 conf.Paths[name].MaxReaders,
			"sourceOnDemand":             conf.Paths[name].SourceOnDemand,
			"sourceOnDemandStartTimeout": conf.Paths[name].SourceOnDemandStartTimeout,
			"sourceOnDemandCloseAfter":   conf.Paths[name].SourceOnDemandCloseAfter,
		}
		data.Items = append(data.Items, pathMap)
	}

	data.ItemCount = len(data.Items)

	// Apply pagination if specified
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}

	return data, nil
}

// GetPathConfig returns the configuration for a specific PathHandler
func (api *MediaMTXAPI) GetPathConfig(name string) (*conf2.Path, error) {
	api.mutex.RLock()
	conf := api.core.Conf
	api.mutex.RUnlock()

	if conf == nil {
		return nil, fmt.Errorf("configuration not available")
	}

	path, exists := conf.Paths[name]
	if !exists {
		return nil, fmt.Errorf("PathHandler '%s' not found", name)
	}

	return path, nil
}

// AddPathConfig adds a new PathHandler configuration
func (api *MediaMTXAPI) AddPathConfig(name string, pathConf *conf2.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	newConf := api.core.Conf.Clone()

	if err := newConf.AddPath(name, pathConf); err != nil {
		return fmt.Errorf("failed to add PathHandler: %v", err)
	}

	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	api.core.Conf = newConf

	// Try to notify Core about config change with timeout to prevent blocking
	select {
	case api.core.ChAPIConfigSet <- newConf:
		// Successfully sent config update
	default:
		// Core event loop not running, update config but skip notification
		// This allows adding paths before Core.Run() is called
	}

	return nil
}

// UpdatePathConfig updates an existing PathHandler configuration (partial update)
func (api *MediaMTXAPI) UpdatePathConfig(name string, pathConf *conf2.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	newConf := api.core.Conf.Clone()

	if err := newConf.PatchPath(name, pathConf); err != nil {
		return fmt.Errorf("failed to update PathHandler: %v", err)
	}

	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	api.core.Conf = newConf

	// Try to notify Core about config change with timeout to prevent blocking
	select {
	case api.core.ChAPIConfigSet <- newConf:
		// Successfully sent config update
	default:
		// Core event loop not running, update config but skip notification
		// This allows updating paths before Core.Run() is called
	}

	return nil
}

// ReplacePathConfig replaces an entire PathHandler configuration
func (api *MediaMTXAPI) ReplacePathConfig(name string, pathConf *conf2.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	newConf := api.core.Conf.Clone()

	if err := newConf.ReplacePath(name, pathConf); err != nil {
		return fmt.Errorf("failed to replace PathHandler: %v", err)
	}

	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	api.core.Conf = newConf

	// Try to notify Core about config change with timeout to prevent blocking
	select {
	case api.core.ChAPIConfigSet <- newConf:
		// Successfully sent config update
	default:
		// Core event loop not running, update config but skip notification
		// This allows replacing paths before Core.Run() is called
	}

	return nil
}

// DeletePathConfig removes a PathHandler configuration
func (api *MediaMTXAPI) DeletePathConfig(name string) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	newConf := api.core.Conf.Clone()

	if err := newConf.RemovePath(name); err != nil {
		return fmt.Errorf("failed to delete PathHandler: %v", err)
	}

	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	api.core.Conf = newConf

	// Try to notify Core about config change with timeout to prevent blocking
	select {
	case api.core.ChAPIConfigSet <- newConf:
		// Successfully sent config update
	default:
		// Core event loop not running, update config but skip notification
		// This allows deleting paths before Core.Run() is called
	}

	return nil
}

// =============================================================================
// RUNTIME PATH INFORMATION
// =============================================================================

// ListActivePaths returns a list of all active paths with pagination
func (api *MediaMTXAPI) ListActivePaths(pagination *PaginationParams) (*defs2.APIPathList, error) {
	data, err := api.core.PathManager.APIPathsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get active paths: %v", err)
	}

	data.ItemCount = len(data.Items)

	// Apply pagination if specified
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}

	return data, nil
}

// GetActivePath returns information about a specific active PathHandler
func (api *MediaMTXAPI) GetActivePath(name string) (*defs2.APIPath, error) {
	data, err := api.core.PathManager.APIPathsGet(name)
	if err != nil {
		if errors.Is(err, conf2.ErrPathNotFound) {
			return nil, fmt.Errorf("PathHandler '%s' not found or not active", name)
		}
		return nil, fmt.Errorf("failed to get PathHandler info: %v", err)
	}

	return data, nil
}

// =============================================================================
// SERVER CONNECTION MANAGEMENT
// =============================================================================

// GetRTSPConnections returns a list of RTSP connections with pagination
func (api *MediaMTXAPI) GetRTSPConnections(pagination *PaginationParams) (*defs2.APIRTSPConnsList, error) {
	if api.core.RtspServer == nil {
		return nil, fmt.Errorf("RTSP server not available")
	}

	data, err := api.core.RtspServer.APIConnsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get RTSP connections: %v", err)
	}

	data.ItemCount = len(data.Items)

	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}

	return data, nil
}

// GetRTSPConnection returns information about a specific RTSP connection
func (api *MediaMTXAPI) GetRTSPConnection(id string) (*defs2.APIRTSPConn, error) {
	if api.core.RtspServer == nil {
		return nil, fmt.Errorf("RTSP server not available")
	}

	connUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %v", err)
	}

	data, err := api.core.RtspServer.APIConnsGet(connUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RTSP connection: %v", err)
	}

	return data, nil
}

// GetRTSPSessions returns a list of RTSP sessions with pagination
func (api *MediaMTXAPI) GetRTSPSessions(pagination *PaginationParams) (*defs2.APIRTSPSessionList, error) {
	if api.core.RtspServer == nil {
		return nil, fmt.Errorf("RTSP server not available")
	}

	data, err := api.core.RtspServer.APISessionsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get RTSP sessions: %v", err)
	}

	data.ItemCount = len(data.Items)

	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}

	return data, nil
}

// GetRTSPSession returns information about a specific RTSP session
func (api *MediaMTXAPI) GetRTSPSession(id string) (*defs2.APIRTSPSession, error) {
	if api.core.RtspServer == nil {
		return nil, fmt.Errorf("RTSP server not available")
	}

	sessionUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %v", err)
	}

	data, err := api.core.RtspServer.APISessionsGet(sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RTSP session: %v", err)
	}

	return data, nil
}

// KickRTSPSession kicks (disconnects) an RTSP session
func (api *MediaMTXAPI) KickRTSPSession(id string) error {
	if api.core.RtspServer == nil {
		return fmt.Errorf("RTSP server not available")
	}

	sessionUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid session ID: %v", err)
	}

	err = api.core.RtspServer.APISessionsKick(sessionUUID)
	if err != nil {
		return fmt.Errorf("failed to kick RTSP session: %v", err)
	}

	return nil
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// sortedPathKeys returns sorted keys from paths map
func (api *MediaMTXAPI) sortedPathKeys(paths map[string]*conf2.Path) []string {
	keys := make([]string, 0, len(paths))
	for name := range paths {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}

// paginateSlice applies pagination to a slice using reflection
func (api *MediaMTXAPI) paginateSlice(itemsPtr interface{}, itemsPerPage, page int) int {
	if itemsPerPage <= 0 {
		return 1
	}

	ritems := reflect.ValueOf(itemsPtr).Elem()
	itemsLen := ritems.Len()

	if itemsLen == 0 {
		return 0
	}

	pageCount := (itemsLen / itemsPerPage)
	if (itemsLen % itemsPerPage) != 0 {
		pageCount++
	}

	minVal := page * itemsPerPage
	if minVal > itemsLen {
		minVal = itemsLen
	}

	maxVal := (page + 1) * itemsPerPage
	if maxVal > itemsLen {
		maxVal = itemsLen
	}

	ritems.Set(ritems.Slice(minVal, maxVal))

	return pageCount
}

// =============================================================================
// CONVENIENCE METHODS
// =============================================================================

// PathOptions represents comprehensive PathHandler configuration options based on OpenAPI PathConf schema
type PathOptions struct {
	Name string `json:"name,omitempty"`

	// General
	Source                     string `json:"source,omitempty"`
	SourceFingerprint          string `json:"sourceFingerprint,omitempty"`
	SourceOnDemand             bool   `json:"sourceOnDemand,omitempty"`
	SourceOnDemandStartTimeout string `json:"sourceOnDemandStartTimeout,omitempty"`
	SourceOnDemandCloseAfter   string `json:"sourceOnDemandCloseAfter,omitempty"`
	MaxReaders                 int    `json:"maxReaders,omitempty"`
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
	RTSPUDPReadBufferSize int    `json:"rtspUDPReadBufferSize,omitempty"`

	// MPEG-TS source
	MPEGTSUDPReadBufferSize int `json:"mpegtsUDPReadBufferSize,omitempty"`

	// RTP source
	RTPSDP               string `json:"rtpSDP,omitempty"`
	RTPUDPReadBufferSize int    `json:"rtpUDPReadBufferSize,omitempty"`

	// Redirect source
	SourceRedirect string `json:"sourceRedirect,omitempty"`

	// Raspberry Pi Camera source
	RPICameraCamID               int       `json:"rpiCameraCamID,omitempty"`
	RPICameraSecondary           bool      `json:"rpiCameraSecondary,omitempty"`
	RPICameraWidth               int       `json:"rpiCameraWidth,omitempty"`
	RPICameraHeight              int       `json:"rpiCameraHeight,omitempty"`
	RPICameraHFlip               bool      `json:"rpiCameraHFlip,omitempty"`
	RPICameraVFlip               bool      `json:"rpiCameraVFlip,omitempty"`
	RPICameraBrightness          float64   `json:"rpiCameraBrightness,omitempty"`
	RPICameraContrast            float64   `json:"rpiCameraContrast,omitempty"`
	RPICameraSaturation          float64   `json:"rpiCameraSaturation,omitempty"`
	RPICameraSharpness           float64   `json:"rpiCameraSharpness,omitempty"`
	RPICameraExposure            string    `json:"rpiCameraExposure,omitempty"`
	RPICameraAWB                 string    `json:"rpiCameraAWB,omitempty"`
	RPICameraAWBGains            []float64 `json:"rpiCameraAWBGains,omitempty"`
	RPICameraDenoise             string    `json:"rpiCameraDenoise,omitempty"`
	RPICameraShutter             int       `json:"rpiCameraShutter,omitempty"`
	RPICameraMetering            string    `json:"rpiCameraMetering,omitempty"`
	RPICameraGain                float64   `json:"rpiCameraGain,omitempty"`
	RPICameraEV                  float64   `json:"rpiCameraEV,omitempty"`
	RPICameraROI                 string    `json:"rpiCameraROI,omitempty"`
	RPICameraHDR                 bool      `json:"rpiCameraHDR,omitempty"`
	RPICameraTuningFile          string    `json:"rpiCameraTuningFile,omitempty"`
	RPICameraMode                string    `json:"rpiCameraMode,omitempty"`
	RPICameraFPS                 float64   `json:"rpiCameraFPS,omitempty"`
	RPICameraAfMode              string    `json:"rpiCameraAfMode,omitempty"`
	RPICameraAfRange             string    `json:"rpiCameraAfRange,omitempty"`
	RPICameraAfSpeed             string    `json:"rpiCameraAfSpeed,omitempty"`
	RPICameraLensPosition        float64   `json:"rpiCameraLensPosition,omitempty"`
	RPICameraAfWindow            string    `json:"rpiCameraAfWindow,omitempty"`
	RPICameraFlickerPeriod       int       `json:"rpiCameraFlickerPeriod,omitempty"`
	RPICameraTextOverlayEnable   bool      `json:"rpiCameraTextOverlayEnable,omitempty"`
	RPICameraTextOverlay         string    `json:"rpiCameraTextOverlay,omitempty"`
	RPICameraCodec               string    `json:"rpiCameraCodec,omitempty"`
	RPICameraIDRPeriod           int       `json:"rpiCameraIDRPeriod,omitempty"`
	RPICameraBitrate             int       `json:"rpiCameraBitrate,omitempty"`
	RPICameraHardwareH264Profile string    `json:"rpiCameraHardwareH264Profile,omitempty"`
	RPICameraHardwareH264Level   string    `json:"rpiCameraHardwareH264Level,omitempty"`
	RPICameraSoftwareH264Profile string    `json:"rpiCameraSoftwareH264Profile,omitempty"`
	RPICameraSoftwareH264Level   string    `json:"rpiCameraSoftwareH264Level,omitempty"`
	RPICameraMJPEGQuality        int       `json:"rpiCameraMJPEGQuality,omitempty"`

	// Hooks
	RunOnInit                  string `json:"runOnInit,omitempty"`
	RunOnInitRestart           bool   `json:"runOnInitRestart,omitempty"`
	RunOnDemand                string `json:"runOnDemand,omitempty"`
	RunOnDemandRestart         bool   `json:"runOnDemandRestart,omitempty"`
	RunOnDemandStartTimeout    string `json:"runOnDemandStartTimeout,omitempty"`
	RunOnDemandCloseAfter      string `json:"runOnDemandCloseAfter,omitempty"`
	RunOnUnDemand              string `json:"runOnUnDemand,omitempty"`
	RunOnReady                 string `json:"runOnReady,omitempty"`
	RunOnReadyRestart          bool   `json:"runOnReadyRestart,omitempty"`
	RunOnNotReady              string `json:"runOnNotReady,omitempty"`
	RunOnRead                  string `json:"runOnRead,omitempty"`
	RunOnReadRestart           bool   `json:"runOnReadRestart,omitempty"`
	RunOnUnread                string `json:"runOnUnread,omitempty"`
	RunOnRecordSegmentCreate   string `json:"runOnRecordSegmentCreate,omitempty"`
	RunOnRecordSegmentComplete string `json:"runOnRecordSegmentComplete,omitempty"`
}

// NewOptionalPath creates a new OptionalPath with the given source
func NewOptionalPath(source string) *conf2.OptionalPath {
	return NewOptionalPathWithOptions(PathOptions{Source: source})
}

// NewOptionalPublisherPath creates a new OptionalPath configured to accept incoming streams
// This is useful when you want to create a PathHandler that accepts streams pushed to it
// rather than pulling from an external source
func NewOptionalPublisherPath(name string) *conf2.OptionalPath {
	return NewOptionalPathWithOptions(PathOptions{
		Name:   name,
		Source: "publisher",
	})
}

// NewOptionalPathWithOptions creates a new OptionalPath with typed options
func NewOptionalPathWithOptions(options PathOptions) *conf2.OptionalPath {
	optPath := &conf2.OptionalPath{}

	// Convert PathOptions to JSON
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		// Fallback to basic publisher PathHandler if marshaling fails
		fallbackJSON := []byte(`{"source":"publisher"}`)
		if options.Source != "" {
			fallbackJSON, _ = json.Marshal(map[string]string{"source": options.Source})
		}
		_ = json.Unmarshal(fallbackJSON, optPath)
		return optPath
	}

	// Use OptionalPath's own UnmarshalJSON method which:
	// 1. Creates the proper dynamic struct with pointer fields
	// 2. Uses jsonwrapper.Unmarshal for proper validation
	// 3. Triggers setDefaults() through the MediaMTX configuration system
	if err := json.Unmarshal(optionsJSON, optPath); err != nil {
		// Fallback to basic publisher PathHandler if unmarshaling fails
		fallbackJSON := []byte(`{"source":"publisher"}`)
		if options.Source != "" {
			fallbackJSON, _ = json.Marshal(map[string]string{"source": options.Source})
		}
		_ = json.Unmarshal(fallbackJSON, optPath)
		return optPath
	}

	return optPath
}

// PaginateFromStrings converts string pagination parameters
func PaginateFromStrings(itemsPerPageStr, pageStr string) (*PaginationParams, error) {
	pagination := DefaultPagination()

	if itemsPerPageStr != "" {
		tmp, err := strconv.ParseUint(itemsPerPageStr, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid itemsPerPage: %v", err)
		}
		if tmp == 0 {
			return nil, fmt.Errorf("itemsPerPage must be greater than 0")
		}
		pagination.ItemsPerPage = int(tmp)
	}

	if pageStr != "" {
		tmp, err := strconv.ParseUint(pageStr, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid page: %v", err)
		}
		pagination.Page = int(tmp)
	}

	return pagination, nil
}

// =============================================================================
// AUTHENTICATION MANAGEMENT
// =============================================================================

// Authenticate performs authentication for API access
func (api *MediaMTXAPI) Authenticate(req *auth.Request) *auth.Error {
	if api.core.AuthManager == nil {
		return nil
	}

	return api.core.AuthManager.Authenticate(req)
}

// RefreshJWTJWKS refreshes JWT JWKS (JSON Web Key Set) configuration
func (api *MediaMTXAPI) RefreshJWTJWKS() {
	if api.core.AuthManager == nil {
		return
	}

	api.core.AuthManager.RefreshJWTJWKS()
}

// CreateAuthRequest creates an authentication request for API access
func (api *MediaMTXAPI) CreateAuthRequest(user, pass, query, ip string) (*auth.Request, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	return &auth.Request{
		Action: conf2.AuthActionAPI,
		Query:  query,
		Credentials: &auth.Credentials{
			User: user,
			Pass: pass,
		},
		IP: parsedIP,
	}, nil
}

// ValidateAPIAccess validates if a user has access to the API
func (api *MediaMTXAPI) ValidateAPIAccess(user, pass, ip string) error {
	req, err := api.CreateAuthRequest(user, pass, "", ip)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %v", err)
	}

	if authErr := api.Authenticate(req); authErr != nil {
		return fmt.Errorf("authentication failed: %v", authErr)
	}

	return nil
}
