package mtx

import (
	"github.com/bluenviron/mediamtx/pkg/auth"
	conf2 "github.com/bluenviron/mediamtx/pkg/conf"
	defs2 "github.com/bluenviron/mediamtx/pkg/defs"
	"github.com/bluenviron/mediamtx/pkg/metrics"
)

// MediaMTXManager is the main interface for managing MediaMTX server operations
type MediaMTXManager interface {
	// Core server operations
	Start() error
	Stop() error
	Restart() error
	IsRunning() bool

	// Configuration management
	GetConfiguration() *conf2.Conf
	UpdateConfiguration(*conf2.Conf) error
	ReloadConfiguration() error

	// Path management
	GetPathCRUDManager() PathCRUDManager
}

// PathCRUDManager provides CRUD operations for PathHandler configurations
// This is the main interface for managing paths programmatically
type PathCRUDManager interface {
	// Core CRUD operations
	ListPaths(itemsPerPage, page int) (*PathConfList, error)
	GetPath(name string) (*conf2.Path, error)
	CreatePath(name string, pathConf *conf2.OptionalPath) error
	UpdatePath(name string, pathConf *conf2.OptionalPath) error
	ReplacePath(name string, pathConf *conf2.OptionalPath) error
	DeletePath(name string) error

	// Validation and defaults
	ValidatePath(name string, pathConf *conf2.OptionalPath) error
	GetPathDefaults() *conf2.Path
	UpdatePathDefaults(pathConf *conf2.OptionalPath) error

	// Active PathHandler information (requires connection to running server)
	GetActivePathsInfo(itemsPerPage, page int) (*defs2.APIPathList, error)
	GetActivePathInfo(name string) (*defs2.APIPath, error)
}

// PathHelper provides utility methods for creating PathHandler configurations
type PathHelperInterface interface {
	CreateBasicPath() *conf2.OptionalPath
	CreateRTSPPath(rtspURL string) *conf2.OptionalPath
	CreateRTMPPath(rtmpURL string) *conf2.OptionalPath
	CreateHLSPath(hlsURL string) *conf2.OptionalPath
	CreateWebRTCPath() *conf2.OptionalPath
	CreateRecordingPath(recordPath string, format conf2.RecordFormat) *conf2.OptionalPath
	CreateOnDemandPath(runOnDemandCmd string) *conf2.OptionalPath
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
	AnalyzePathSource(path *conf2.Path) map[string]interface{}
	AnalyzePathStatus(apiPath *defs2.APIPath) map[string]interface{}
}

// PathQuery provides query and filtering capabilities for paths
type PathQueryInterface interface {
	FilterPathsBySource(paths []*conf2.Path, sourceType string) []*conf2.Path
	FilterPathsByRecording(paths []*conf2.Path, recordingEnabled bool) []*conf2.Path
	FilterPathsByOnDemand(paths []*conf2.Path, onDemandEnabled bool) []*conf2.Path
	SearchPathsByName(pathNames []string, pattern string) []string
}

// PathStats provides statistics and metrics for paths
type PathStatsInterface interface {
	CalculateTrafficStats(apiPaths []*defs2.APIPath) map[string]interface{}
	CountPathsByType(paths []*conf2.Path) map[string]int
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
	ListRTSPConnections() ([]*defs2.APIRTSPConn, error)
	ListRTMPConnections() ([]*defs2.APIRTMPConn, error)
	ListSRTConnections() ([]*defs2.APISRTConn, error)
	ListWebRTCConnections() ([]*defs2.APIWebRTCSession, error)

	// Connection management
	DisconnectRTSPConnection(id string) error
	DisconnectRTMPConnection(id string) error
	DisconnectSRTConnection(id string) error
	DisconnectWebRTCConnection(id string) error

	// Session management
	ListRTSPSessions() ([]*defs2.APIRTSPSession, error)
	DisconnectRTSPSession(id string) error
}

// StreamManager provides stream management operations
type StreamManager interface {
	// Stream information
	GetStreamInfo(pathName string) (*defs2.APIPath, error)
	ListActiveStreams() ([]*defs2.APIPath, error)

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
	ListRecordings(pathName string) ([]*defs2.APIRecording, error)
	GetRecordingInfo(pathName, fileName string) (*defs2.APIRecordingSegment, error)
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
	ValidateConfiguration(config *conf2.Conf) error
	GetConfigurationErrors() []error

	// Configuration sections
	GetGlobalConfig() *conf2.Global
	UpdateGlobalConfig(global *conf2.Global) error

	// Configuration templates
	GetDefaultConfiguration() *conf2.Conf
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
func NewStreamManager(pathManager defs2.APIPathManager) StreamManager {
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
