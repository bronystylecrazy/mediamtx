package mediamtx

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strconv"
	"sync"

	"github.com/google/uuid"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/auth"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/defs"
)

// DirectAPI provides programmatic access to MediaMTX functionality without HTTP/gin dependencies
type DirectAPI struct {
	core        *Core
	mutex       sync.RWMutex
}

// NewDirectAPI creates a new direct API instance
func NewDirectAPI(core *Core) *DirectAPI {
	return &DirectAPI{
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

// DirectAPIPathConfList represents a list of path configurations for direct API
type DirectAPIPathConfList struct {
	ItemCount int                  `json:"itemCount"`
	PageCount int                  `json:"pageCount"`
	Items     []map[string]interface{} `json:"items"`
}

// =============================================================================
// CONFIGURATION MANAGEMENT
// =============================================================================

// GetGlobalConfig returns the current global configuration
func (api *DirectAPI) GetGlobalConfig() (*conf.Conf, error) {
	api.mutex.RLock()
	defer api.mutex.RUnlock()
	
	if api.core.Conf == nil {
		return nil, fmt.Errorf("configuration not available")
	}
	return api.core.Conf.Clone(), nil
}

// UpdateGlobalConfig updates the global configuration
func (api *DirectAPI) UpdateGlobalConfig(newConf *conf.Conf) error {
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
func (api *DirectAPI) PatchGlobalConfig(optionalGlobal *conf.OptionalGlobal) error {
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

// GetPathDefaults returns the default path configuration
func (api *DirectAPI) GetPathDefaults() *conf.Path {
	return &conf.Path{}
}

// UpdatePathDefaults updates the default path configuration
func (api *DirectAPI) UpdatePathDefaults(defaults *conf.OptionalPath) error {
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

// PatchPathDefaults patches the default path configuration (equivalent to PATCH /config/pathdefaults/patch)
func (api *DirectAPI) PatchPathDefaults(optionalPath *conf.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()
	
	newConf := api.core.Conf.Clone()
	
	// Apply the patch to path defaults
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
func (api *DirectAPI) ListPathConfigs(pagination *PaginationParams) (*DirectAPIPathConfList, error) {
	api.mutex.RLock()
	conf := api.core.Conf
	api.mutex.RUnlock()
	
	if conf == nil {
		return nil, fmt.Errorf("configuration not available")
	}
	
	// Create sorted list of paths
	sortedNames := api.sortedPathKeys(conf.Paths)
	data := &DirectAPIPathConfList{
		Items: []map[string]interface{}{},
	}
	
	for _, name := range sortedNames {
		// Convert path to map for JSON serialization
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

// GetPathConfig returns the configuration for a specific path
func (api *DirectAPI) GetPathConfig(name string) (*conf.Path, error) {
	api.mutex.RLock()
	conf := api.core.Conf
	api.mutex.RUnlock()
	
	if conf == nil {
		return nil, fmt.Errorf("configuration not available")
	}
	
	path, exists := conf.Paths[name]
	if !exists {
		return nil, fmt.Errorf("path '%s' not found", name)
	}
	
	return path, nil
}

// AddPathConfig adds a new path configuration
func (api *DirectAPI) AddPathConfig(name string, pathConf *conf.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()
	
	newConf := api.core.Conf.Clone()
	
	if err := newConf.AddPath(name, pathConf); err != nil {
		return fmt.Errorf("failed to add path: %v", err)
	}
	
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)
	
	return nil
}

// UpdatePathConfig updates an existing path configuration (partial update)
func (api *DirectAPI) UpdatePathConfig(name string, pathConf *conf.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()
	
	newConf := api.core.Conf.Clone()
	
	if err := newConf.PatchPath(name, pathConf); err != nil {
		return fmt.Errorf("failed to update path: %v", err)
	}
	
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)
	
	return nil
}

// ReplacePathConfig replaces an entire path configuration
func (api *DirectAPI) ReplacePathConfig(name string, pathConf *conf.OptionalPath) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()
	
	newConf := api.core.Conf.Clone()
	
	if err := newConf.ReplacePath(name, pathConf); err != nil {
		return fmt.Errorf("failed to replace path: %v", err)
	}
	
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)
	
	return nil
}

// DeletePathConfig removes a path configuration
func (api *DirectAPI) DeletePathConfig(name string) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()
	
	newConf := api.core.Conf.Clone()
	
	if err := newConf.RemovePath(name); err != nil {
		return fmt.Errorf("failed to delete path: %v", err)
	}
	
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)
	
	return nil
}

// =============================================================================
// RUNTIME PATH INFORMATION
// =============================================================================

// ListActivePaths returns a list of all active paths with pagination
func (api *DirectAPI) ListActivePaths(pagination *PaginationParams) (*defs.APIPathList, error) {
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

// GetActivePath returns information about a specific active path
func (api *DirectAPI) GetActivePath(name string) (*defs.APIPath, error) {
	data, err := api.core.PathManager.APIPathsGet(name)
	if err != nil {
		if errors.Is(err, conf.ErrPathNotFound) {
			return nil, fmt.Errorf("path '%s' not found or not active", name)
		}
		return nil, fmt.Errorf("failed to get path info: %v", err)
	}
	
	return data, nil
}

// =============================================================================
// SERVER CONNECTION MANAGEMENT  
// =============================================================================

// GetRTSPConnections returns a list of RTSP connections with pagination
func (api *DirectAPI) GetRTSPConnections(pagination *PaginationParams) (*defs.APIRTSPConnsList, error) {
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
func (api *DirectAPI) GetRTSPConnection(id string) (*defs.APIRTSPConn, error) {
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
func (api *DirectAPI) GetRTSPSessions(pagination *PaginationParams) (*defs.APIRTSPSessionList, error) {
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
func (api *DirectAPI) GetRTSPSession(id string) (*defs.APIRTSPSession, error) {
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
func (api *DirectAPI) KickRTSPSession(id string) error {
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
func (api *DirectAPI) sortedPathKeys(paths map[string]*conf.Path) []string {
	keys := make([]string, 0, len(paths))
	for name := range paths {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}

// paginateSlice applies pagination to a slice using reflection
func (api *DirectAPI) paginateSlice(itemsPtr interface{}, itemsPerPage, page int) int {
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
func (api *DirectAPI) Authenticate(req *auth.Request) *auth.Error {
	if api.core.AuthManager == nil {
		return nil
	}
	
	return api.core.AuthManager.Authenticate(req)
}

// RefreshJWTJWKS refreshes JWT JWKS (JSON Web Key Set) configuration
func (api *DirectAPI) RefreshJWTJWKS() {
	if api.core.AuthManager == nil {
		return
	}
	
	api.core.AuthManager.RefreshJWTJWKS()
}

// CreateAuthRequest creates an authentication request for API access
func (api *DirectAPI) CreateAuthRequest(user, pass, query, ip string) (*auth.Request, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}
	
	return &auth.Request{
		Action:      conf.AuthActionAPI,
		Query:       query,
		Credentials: &auth.Credentials{
			User: user,
			Pass: pass,
		},
		IP: parsedIP,
	}, nil
}

// ValidateAPIAccess validates if a user has access to the API
func (api *DirectAPI) ValidateAPIAccess(user, pass, ip string) error {
	req, err := api.CreateAuthRequest(user, pass, "", ip)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %v", err)
	}
	
	if authErr := api.Authenticate(req); authErr != nil {
		return fmt.Errorf("authentication failed: %v", authErr)
	}
	
	return nil
}