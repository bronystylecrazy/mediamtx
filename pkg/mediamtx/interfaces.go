package mediamtx

import (
	"github.com/bluenviron/mediamtx/pkg/mediamtx/auth"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/defs"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/metrics"
)

// MediaMTXManager is the main interface for managing MediaMTX server operations
type MediaMTXManager interface {
	// Core server operations
	Start() error
	Stop() error
	Restart() error
	IsRunning() bool

	// Configuration management
	GetConfiguration() *conf.Conf
	UpdateConfiguration(*conf.Conf) error
	ReloadConfiguration() error
	
	// Path management
	GetPathCRUDManager() PathCRUDManager
}

// PathCRUDManager provides CRUD operations for PathHandler configurations
// This is the main interface for managing paths programmatically
type PathCRUDManager interface {
	// Core CRUD operations
	ListPaths(itemsPerPage, page int) (*PathConfList, error)
	GetPath(name string) (*conf.Path, error)
	CreatePath(name string, pathConf *conf.OptionalPath) error
	UpdatePath(name string, pathConf *conf.OptionalPath) error
	ReplacePath(name string, pathConf *conf.OptionalPath) error
	DeletePath(name string) error

	// Validation and defaults
	ValidatePath(name string, pathConf *conf.OptionalPath) error
	GetPathDefaults() *conf.Path
	UpdatePathDefaults(pathConf *conf.OptionalPath) error

	// Active PathHandler information (requires connection to running server)
	GetActivePathsInfo(itemsPerPage, page int) (*defs.APIPathList, error)
	GetActivePathInfo(name string) (*defs.APIPath, error)
}

// PathHelper provides utility methods for creating PathHandler configurations
type PathHelperInterface interface {
	CreateBasicPath() *conf.OptionalPath
	CreateRTSPPath(rtspURL string) *conf.OptionalPath
	CreateRTMPPath(rtmpURL string) *conf.OptionalPath
	CreateHLSPath(hlsURL string) *conf.OptionalPath
	CreateWebRTCPath() *conf.OptionalPath
	CreateRecordingPath(recordPath string, format conf.RecordFormat) *conf.OptionalPath
	CreateOnDemandPath(runOnDemandCmd string) *conf.OptionalPath
}

// PathValidator provides validation methods for PathHandler configurations
type PathValidatorInterface interface {
	ValidatePathName(name string) error
	ValidateRTSPURL(url string) error
	ValidateRTMPURL(url string) error
	ValidateHLSURL(url string) error
}

// PathAnalyzer provides analysis methods for PathHandler configurations and status
type PathAnalyzerInterface interface {
	AnalyzePathSource(path *conf.Path) map[string]interface{}
	AnalyzePathStatus(apiPath *defs.APIPath) map[string]interface{}
}

// PathQuery provides query and filtering capabilities for paths
type PathQueryInterface interface {
	FilterPathsBySource(paths []*conf.Path, sourceType string) []*conf.Path
	FilterPathsByRecording(paths []*conf.Path, recordingEnabled bool) []*conf.Path
	FilterPathsByOnDemand(paths []*conf.Path, onDemandEnabled bool) []*conf.Path
	SearchPathsByName(pathNames []string, pattern string) []string
}

// PathStats provides statistics and metrics for paths
type PathStatsInterface interface {
	CalculateTrafficStats(apiPaths []*defs.APIPath) map[string]interface{}
	CountPathsByType(paths []*conf.Path) map[string]int
}

// ServerManager provides server-level management operations
type ServerManager interface {
	// Server lifecycle
	Initialize() error
	Start() error
	Stop() error
	Restart() error
	IsRunning() bool
	GetStatus() map[string]interface{}

	// Server configuration
	GetListenAddresses() map[string]string
	GetVersion() string
	GetUptime() string

	// Protocol server management
	EnableProtocol(protocol string) error
	DisableProtocol(protocol string) error
	GetEnabledProtocols() []string
}

// ConnectionManager provides connection management operations
type ConnectionManager interface {
	// List connections by protocol
	ListRTSPConnections() ([]*defs.APIRTSPConn, error)
	ListRTMPConnections() ([]*defs.APIRTMPConn, error)
	ListSRTConnections() ([]*defs.APISRTConn, error)
	ListWebRTCConnections() ([]*defs.APIWebRTCSession, error)

	// Connection management
	DisconnectRTSPConnection(id string) error
	DisconnectRTMPConnection(id string) error
	DisconnectSRTConnection(id string) error
	DisconnectWebRTCConnection(id string) error

	// Session management
	ListRTSPSessions() ([]*defs.APIRTSPSession, error)
	DisconnectRTSPSession(id string) error
}

// StreamManager provides stream management operations
type StreamManager interface {
	// Stream information
	GetStreamInfo(pathName string) (*defs.APIPath, error)
	ListActiveStreams() ([]*defs.APIPath, error)

	// Stream control
	StartStream(pathName string) error
	StopStream(pathName string) error
	RestartStream(pathName string) error

	// Stream statistics
	GetStreamStats(pathName string) (map[string]interface{}, error)
	GetGlobalStreamStats() (map[string]interface{}, error)
}

// RecordingManager provides recording management operations
type RecordingManager interface {
	// Recording control
	StartRecording(pathName string) error
	StopRecording(pathName string) error
	IsRecording(pathName string) (bool, error)

	// Recording file management
	ListRecordings(pathName string) ([]*defs.APIRecording, error)
	GetRecordingInfo(pathName, fileName string) (*defs.APIRecordingSegment, error)
	DeleteRecording(pathName, fileName string) error

	// Recording configuration
	GetRecordingPath(pathName string) (string, error)
	SetRecordingPath(pathName, recordPath string) error
}

// AuthManager provides authentication and authorization management
type AuthManager interface {
	// User management
	AddUser(username, password string, permissions []string) error
	UpdateUser(username, password string, permissions []string) error
	DeleteUser(username string) error
	ListUsers() ([]string, error)

	// Permission management
	GrantPermission(username, pathPattern, action string) error
	RevokePermission(username, pathPattern, action string) error
	CheckPermission(username, pathPattern, action string) (bool, error)

	// Authentication methods
	SetAuthMethod(method string) error
	GetAuthMethod() string
	RefreshJWTKeys() error
}

// MetricsManager provides metrics and monitoring capabilities
type MetricsManager interface {
	// System metrics
	GetSystemMetrics() (map[string]interface{}, error)
	GetMemoryUsage() (map[string]interface{}, error)
	GetCPUUsage() (float64, error)

	// Traffic metrics
	GetTrafficMetrics() (map[string]interface{}, error)
	GetProtocolMetrics(protocol string) (map[string]interface{}, error)

	// Path metrics
	GetPathMetrics(pathName string) (map[string]interface{}, error)
	GetTopPaths(limit int) ([]map[string]interface{}, error)

	// Export metrics
	ExportPrometheusMetrics() (string, error)
	ExportJSONMetrics() (map[string]interface{}, error)
}

// ConfigurationManager provides configuration management operations
type ConfigurationManager interface {
	// Configuration file operations
	LoadFromFile(filePath string) error
	SaveToFile(filePath string) error
	GetConfigPath() string

	// Configuration validation
	ValidateConfiguration(config *conf.Conf) error
	GetConfigurationErrors() []error

	// Configuration sections
	GetGlobalConfig() *conf.Global
	UpdateGlobalConfig(global *conf.Global) error

	// Configuration templates
	GetDefaultConfiguration() *conf.Conf
	CreateConfigurationTemplate() map[string]interface{}
}

// EventManager provides event handling and notification capabilities
type EventManager interface {
	// Event subscription
	Subscribe(eventType string, handler func(interface{})) error
	Unsubscribe(eventType string, handler func(interface{})) error

	// Event publishing
	PublishEvent(eventType string, data interface{}) error

	// Event types
	GetSupportedEventTypes() []string

	// Event history
	GetRecentEvents(limit int) ([]interface{}, error)
	ClearEventHistory() error
}

// Factory functions for creating managers

// NewMediaMTXManager creates a new MediaMTX manager instance
func NewMediaMTXManager(configPath string) (MediaMTXManager, error) {
	// Implementation would go here
	return nil, nil
}

// NewServerManager creates a new server manager instance
func NewServerManager(core *Core) ServerManager {
	// Implementation would go here
	return nil
}

// NewConnectionManager creates a new connection manager instance
func NewConnectionManager(core *Core) ConnectionManager {
	// Implementation would go here
	return nil
}

// NewStreamManager creates a new stream manager instance
func NewStreamManager(pathManager defs.APIPathManager) StreamManager {
	// Implementation would go here
	return nil
}

// NewRecordingManager creates a new recording manager instance
func NewRecordingManager(core *Core) RecordingManager {
	// Implementation would go here
	return nil
}

// NewAuthManager creates a new authentication manager instance
func NewAuthManager(authMgr *auth.Manager) AuthManager {
	// Implementation would go here
	return nil
}

// NewMetricsManager creates a new metrics manager instance  
func NewMetricsManager(metrics *metrics.Metrics) MetricsManager {
	// Implementation would go here
	return nil
}

// NewConfigurationManager creates a new configuration manager instance
func NewConfigurationManager() ConfigurationManager {
	// Implementation would go here
	return nil
}

// NewEventManager creates a new event manager instance
func NewEventManager() EventManager {
	// Implementation would go here
	return nil
}
