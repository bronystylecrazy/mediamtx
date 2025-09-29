package mediamtx

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/defs"
)

// PathCRUDManager interface is defined in interfaces.go to avoid duplication

// PathCRUDError represents errors that can occur during PathHandler CRUD operations
type PathCRUDError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e PathCRUDError) Error() string {
	return e.Message
}

// Common error types
var (
	ErrPathNotFound     = &PathCRUDError{Type: "path_not_found", Message: "PathHandler configuration not found", Code: 404}
	ErrPathExists       = &PathCRUDError{Type: "path_exists", Message: "PathHandler configuration already exists", Code: 409}
	ErrInvalidPathName  = &PathCRUDError{Type: "invalid_path_name", Message: "invalid PathHandler name", Code: 400}
	ErrValidationFailed = &PathCRUDError{Type: "validation_failed", Message: "PathHandler configuration validation failed", Code: 400}
	ErrInternalError    = &PathCRUDError{Type: "internal_error", Message: "internal server error", Code: 500}
)

// pathCRUDManager implements PathCRUDManager interface
type pathCRUDManager struct {
	mutex       sync.RWMutex
	conf        *conf.Conf
	pathManager defs.APIPathManager
	onConfigSet func(*conf.Conf)
}

// NewPathCRUDManager creates a new PathCRUDManager instance
func NewPathCRUDManager(config *conf.Conf, pathManager defs.APIPathManager, onConfigSet func(*conf.Conf)) PathCRUDManager {
	return &pathCRUDManager{
		conf:        config,
		pathManager: pathManager,
		onConfigSet: onConfigSet,
	}
}

// ListPaths returns all PathHandler configurations with pagination support
func (m *pathCRUDManager) ListPaths(itemsPerPage, page int) (*PathConfList, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	data := &PathConfList{
		Items: make([]PathConf, 0, len(m.conf.Paths)),
	}

	// Sort paths by name
	sortedNames := make([]string, 0, len(m.conf.Paths))
	for name := range m.conf.Paths {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	// Convert from conf.Path to PathConf
	for _, name := range sortedNames {
		confPath := m.conf.Paths[name]
		// Convert to JSON and back to get PathConf
		jsonData, err := json.Marshal(confPath)
		if err != nil {
			continue // Skip invalid paths
		}

		var pathConf PathConf
		if err := json.Unmarshal(jsonData, &pathConf); err != nil {
			continue // Skip invalid paths
		}

		pathConf.Name = name
		data.Items = append(data.Items, pathConf)
	}

	data.ItemCount = int64(len(data.Items))

	// Apply pagination if requested
	if itemsPerPage > 0 {
		pageCount, err := m.paginatePathConf(&data.Items, itemsPerPage, page)
		if err != nil {
			return nil, &PathCRUDError{
				Type:    "pagination_error",
				Message: fmt.Sprintf("pagination failed: %v", err),
				Code:    400,
			}
		}
		data.PageCount = int64(pageCount)
	} else {
		data.PageCount = 1
	}

	return data, nil
}

// GetPath returns a specific PathHandler configuration by name
func (m *pathCRUDManager) GetPath(name string) (*conf.Path, error) {
	if err := conf.IsValidPathName(name); err != nil {
		return nil, &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	path, exists := m.conf.Paths[name]
	if !exists {
		return nil, &PathCRUDError{
			Type:    ErrPathNotFound.Type,
			Message: fmt.Sprintf("PathHandler '%s' not found", name),
			Code:    ErrPathNotFound.Code,
		}
	}

	return path, nil
}

// CreatePath creates a new PathHandler configuration
func (m *pathCRUDManager) CreatePath(name string, pathConf *conf.OptionalPath) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if PathHandler already exists
	if _, exists := m.conf.Paths[name]; exists {
		return &PathCRUDError{
			Type:    ErrPathExists.Type,
			Message: fmt.Sprintf("PathHandler '%s' already exists", name),
			Code:    ErrPathExists.Code,
		}
	}

	newConf := m.conf.Clone()

	if err := newConf.AddPath(name, pathConf); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("failed to add PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	m.conf = newConf
	if m.onConfigSet != nil {
		go m.onConfigSet(newConf)
	}

	return nil
}

// UpdatePath partially updates an existing PathHandler configuration
func (m *pathCRUDManager) UpdatePath(name string, pathConf *conf.OptionalPath) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	newConf := m.conf.Clone()

	if err := newConf.PatchPath(name, pathConf); err != nil {
		if errors.Is(err, conf.ErrPathNotFound) {
			return &PathCRUDError{
				Type:    ErrPathNotFound.Type,
				Message: fmt.Sprintf("PathHandler '%s' not found", name),
				Code:    ErrPathNotFound.Code,
			}
		}
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("failed to update PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	m.conf = newConf
	if m.onConfigSet != nil {
		go m.onConfigSet(newConf)
	}

	return nil
}

// ReplacePath completely replaces an existing PathHandler configuration
func (m *pathCRUDManager) ReplacePath(name string, pathConf *conf.OptionalPath) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	newConf := m.conf.Clone()

	if err := newConf.ReplacePath(name, pathConf); err != nil {
		if errors.Is(err, conf.ErrPathNotFound) {
			return &PathCRUDError{
				Type:    ErrPathNotFound.Type,
				Message: fmt.Sprintf("PathHandler '%s' not found", name),
				Code:    ErrPathNotFound.Code,
			}
		}
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("failed to replace PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	m.conf = newConf
	if m.onConfigSet != nil {
		go m.onConfigSet(newConf)
	}

	return nil
}

// DeletePath removes a PathHandler configuration
func (m *pathCRUDManager) DeletePath(name string) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	newConf := m.conf.Clone()

	if err := newConf.RemovePath(name); err != nil {
		if errors.Is(err, conf.ErrPathNotFound) {
			return &PathCRUDError{
				Type:    ErrPathNotFound.Type,
				Message: fmt.Sprintf("PathHandler '%s' not found", name),
				Code:    ErrPathNotFound.Code,
			}
		}
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("failed to remove PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed after removing PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	m.conf = newConf
	if m.onConfigSet != nil {
		go m.onConfigSet(newConf)
	}

	return nil
}

// ValidatePath validates a PathHandler configuration without saving it
func (m *pathCRUDManager) ValidatePath(name string, pathConf *conf.OptionalPath) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a temporary config for validation
	tempConf := m.conf.Clone()

	// Try to add/update the PathHandler temporarily
	if _, exists := tempConf.Paths[name]; exists {
		if err := tempConf.PatchPath(name, pathConf); err != nil {
			return &PathCRUDError{
				Type:    ErrValidationFailed.Type,
				Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
				Code:    ErrValidationFailed.Code,
			}
		}
	} else {
		if err := tempConf.AddPath(name, pathConf); err != nil {
			return &PathCRUDError{
				Type:    ErrValidationFailed.Type,
				Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
				Code:    ErrValidationFailed.Code,
			}
		}
	}

	if err := tempConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	return nil
}

// GetPathDefaults returns the default PathHandler configuration
func (m *pathCRUDManager) GetPathDefaults() *conf.Path {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return &m.conf.PathDefaults
}

// UpdatePathDefaults updates the default PathHandler configuration
func (m *pathCRUDManager) UpdatePathDefaults(pathConf *conf.OptionalPath) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	newConf := m.conf.Clone()
	newConf.PatchPathDefaults(pathConf)

	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler defaults: %v", err),
			Code:    ErrValidationFailed.Code,
		}
	}

	m.conf = newConf
	if m.onConfigSet != nil {
		go m.onConfigSet(newConf)
	}

	return nil
}

// GetActivePathsInfo returns information about currently active paths
func (m *pathCRUDManager) GetActivePathsInfo(itemsPerPage, page int) (*defs.APIPathList, error) {
	if m.pathManager == nil {
		return nil, &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "PathHandler manager not available",
			Code:    ErrInternalError.Code,
		}
	}

	data, err := m.pathManager.APIPathsList()
	if err != nil {
		return nil, &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: fmt.Sprintf("failed to get active paths: %v", err),
			Code:    ErrInternalError.Code,
		}
	}

	// Apply pagination if requested
	if itemsPerPage > 0 {
		pageCount, err := m.paginateAPIPaths(&data.Items, itemsPerPage, page)
		if err != nil {
			return nil, &PathCRUDError{
				Type:    "pagination_error",
				Message: fmt.Sprintf("pagination failed: %v", err),
				Code:    400,
			}
		}
		data.PageCount = pageCount
	}

	return data, nil
}

// GetActivePathInfo returns information about a specific active PathHandler
func (m *pathCRUDManager) GetActivePathInfo(name string) (*defs.APIPath, error) {
	if err := conf.IsValidPathName(name); err != nil {
		return nil, &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	if m.pathManager == nil {
		return nil, &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "PathHandler manager not available",
			Code:    ErrInternalError.Code,
		}
	}

	path, err := m.pathManager.APIPathsGet(name)
	if err != nil {
		return nil, &PathCRUDError{
			Type:    ErrPathNotFound.Type,
			Message: fmt.Sprintf("active PathHandler '%s' not found: %v", name, err),
			Code:    ErrPathNotFound.Code,
		}
	}

	return path, nil
}

// Helper method for pagination
func (m *pathCRUDManager) paginate(items *[]*conf.Path, itemsPerPage, page int) (int, error) {
	if itemsPerPage <= 0 {
		return 1, nil
	}

	if page < 1 {
		page = 1
	}

	totalItems := len(*items)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	if page > totalPages && totalPages > 0 {
		return 0, fmt.Errorf("page %d out of range (total pages: %d)", page, totalPages)
	}

	startIdx := (page - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage

	if startIdx >= totalItems {
		*items = []*conf.Path{}
	} else {
		if endIdx > totalItems {
			endIdx = totalItems
		}
		*items = (*items)[startIdx:endIdx]
	}

	return totalPages, nil
}

// paginatePathConf applies pagination to a slice of PathConf
func (m *pathCRUDManager) paginatePathConf(items *[]PathConf, itemsPerPage, page int) (int, error) {
	if itemsPerPage <= 0 {
		return 1, nil
	}

	if page < 1 {
		page = 1
	}

	totalItems := len(*items)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	if page > totalPages && totalPages > 0 {
		return 0, fmt.Errorf("page %d out of range (total pages: %d)", page, totalPages)
	}

	startIdx := (page - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage

	if startIdx >= totalItems {
		*items = []PathConf{}
	} else {
		if endIdx > totalItems {
			endIdx = totalItems
		}
		*items = (*items)[startIdx:endIdx]
	}

	return totalPages, nil
}

// Helper method for API PathHandler pagination
func (m *pathCRUDManager) paginateAPIPaths(items *[]*defs.APIPath, itemsPerPage, page int) (int, error) {
	if itemsPerPage <= 0 {
		return 1, nil
	}

	if page < 1 {
		page = 1
	}

	totalItems := len(*items)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	if page > totalPages && totalPages > 0 {
		return 0, fmt.Errorf("page %d out of range (total pages: %d)", page, totalPages)
	}

	startIdx := (page - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage

	if startIdx >= totalItems {
		*items = []*defs.APIPath{}
	} else {
		if endIdx > totalItems {
			endIdx = totalItems
		}
		*items = (*items)[startIdx:endIdx]
	}

	return totalPages, nil
}
