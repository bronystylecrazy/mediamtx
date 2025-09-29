package mediamtx

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/defs"
)

// PathCRUDManager implementation that integrates with Core
type corePathCRUDManager struct {
	core *Core
	mu   sync.RWMutex
}

// GetPathCRUDManager returns a PathCRUDManager that operates on the Core instance
func (p *Core) GetPathCRUDManager() PathCRUDManager {
	return &corePathCRUDManager{
		core: p,
	}
}

// ListPaths returns all PathHandler configurations with pagination support
func (m *corePathCRUDManager) ListPaths(itemsPerPage, page int) (*PathConfList, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.core.Conf == nil {
		return nil, &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "core configuration not available",
			Code:    ErrInternalError.Code,
		}
	}

	data := &PathConfList{
		Items: make([]PathConf, 0, len(m.core.Conf.Paths)),
	}

	// Sort paths by name and convert to PathConf
	for name, path := range m.core.Conf.Paths {
		// Convert from conf.Path to PathConf
		jsonData, err := json.Marshal(path)
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
func (m *corePathCRUDManager) GetPath(name string) (*conf.Path, error) {
	if err := conf.IsValidPathName(name); err != nil {
		return nil, &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.core.Conf == nil {
		return nil, &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "core configuration not available",
			Code:    ErrInternalError.Code,
		}
	}

	path, exists := m.core.Conf.Paths[name]
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
func (m *corePathCRUDManager) CreatePath(name string, pathConf *conf.OptionalPath) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.core.Conf == nil {
		return &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "core configuration not available",
			Code:    ErrInternalError.Code,
		}
	}

	// Check if PathHandler already exists
	if _, exists := m.core.Conf.Paths[name]; exists {
		return &PathCRUDError{
			Type:    ErrPathExists.Type,
			Message: fmt.Sprintf("PathHandler '%s' already exists", name),
			Code:    ErrPathExists.Code,
		}
	}

	// Clone the current configuration
	newConf := m.core.Conf.Clone()

	// Add the new PathHandler
	if err := newConf.AddPath(name, pathConf); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("failed to add PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	// Validate the new configuration
	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	// Apply the configuration through Core's API config channel
	m.core.APIConfigSet(newConf)

	return nil
}

// UpdatePath partially updates an existing PathHandler configuration
func (m *corePathCRUDManager) UpdatePath(name string, pathConf *conf.OptionalPath) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.core.Conf == nil {
		return &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "core configuration not available",
			Code:    ErrInternalError.Code,
		}
	}

	// Clone the current configuration
	newConf := m.core.Conf.Clone()

	// Update the PathHandler
	if err := newConf.PatchPath(name, pathConf); err != nil {
		if err == conf.ErrPathNotFound {
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

	// Validate the new configuration
	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	// Apply the configuration through Core's API config channel
	m.core.APIConfigSet(newConf)

	return nil
}

// ReplacePath completely replaces an existing PathHandler configuration
func (m *corePathCRUDManager) ReplacePath(name string, pathConf *conf.OptionalPath) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.core.Conf == nil {
		return &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "core configuration not available",
			Code:    ErrInternalError.Code,
		}
	}

	// Clone the current configuration
	newConf := m.core.Conf.Clone()

	// Replace the PathHandler
	if err := newConf.ReplacePath(name, pathConf); err != nil {
		if err == conf.ErrPathNotFound {
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

	// Validate the new configuration
	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	// Apply the configuration through Core's API config channel
	m.core.APIConfigSet(newConf)

	return nil
}

// DeletePath removes a PathHandler configuration
func (m *corePathCRUDManager) DeletePath(name string) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.core.Conf == nil {
		return &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "core configuration not available",
			Code:    ErrInternalError.Code,
		}
	}

	// Clone the current configuration
	newConf := m.core.Conf.Clone()

	// Remove the PathHandler
	if err := newConf.RemovePath(name); err != nil {
		if err == conf.ErrPathNotFound {
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

	// Validate the new configuration
	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed after removing PathHandler '%s': %v", name, err),
			Code:    ErrValidationFailed.Code,
		}
	}

	// Apply the configuration through Core's API config channel
	m.core.APIConfigSet(newConf)

	return nil
}

// ValidatePath validates a PathHandler configuration without saving it
func (m *corePathCRUDManager) ValidatePath(name string, pathConf *conf.OptionalPath) error {
	if err := conf.IsValidPathName(name); err != nil {
		return &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.core.Conf == nil {
		return &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "core configuration not available",
			Code:    ErrInternalError.Code,
		}
	}

	// Create a temporary config for validation
	tempConf := m.core.Conf.Clone()

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
func (m *corePathCRUDManager) GetPathDefaults() *conf.Path {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.core.Conf == nil {
		return &conf.Path{} // Return empty PathHandler if no config
	}

	return &m.core.Conf.PathDefaults
}

// UpdatePathDefaults updates the default PathHandler configuration
func (m *corePathCRUDManager) UpdatePathDefaults(pathConf *conf.OptionalPath) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.core.Conf == nil {
		return &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "core configuration not available",
			Code:    ErrInternalError.Code,
		}
	}

	// Clone the current configuration
	newConf := m.core.Conf.Clone()

	// Update PathHandler defaults
	newConf.PatchPathDefaults(pathConf)

	// Validate the new configuration
	if err := newConf.Validate(nil); err != nil {
		return &PathCRUDError{
			Type:    ErrValidationFailed.Type,
			Message: fmt.Sprintf("validation failed for PathHandler defaults: %v", err),
			Code:    ErrValidationFailed.Code,
		}
	}

	// Apply the configuration through Core's API config channel
	m.core.APIConfigSet(newConf)

	return nil
}

// GetActivePathsInfo returns information about currently active paths
func (m *corePathCRUDManager) GetActivePathsInfo(itemsPerPage, page int) (*defs.APIPathList, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.core.PathManager == nil {
		return nil, &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "PathHandler manager not available",
			Code:    ErrInternalError.Code,
		}
	}

	data, err := m.core.PathManager.APIPathsList()
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
func (m *corePathCRUDManager) GetActivePathInfo(name string) (*defs.APIPath, error) {
	if err := conf.IsValidPathName(name); err != nil {
		return nil, &PathCRUDError{
			Type:    ErrInvalidPathName.Type,
			Message: fmt.Sprintf("invalid PathHandler name '%s': %v", name, err),
			Code:    ErrInvalidPathName.Code,
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.core.PathManager == nil {
		return nil, &PathCRUDError{
			Type:    ErrInternalError.Type,
			Message: "PathHandler manager not available",
			Code:    ErrInternalError.Code,
		}
	}

	path, err := m.core.PathManager.APIPathsGet(name)
	if err != nil {
		return nil, &PathCRUDError{
			Type:    ErrPathNotFound.Type,
			Message: fmt.Sprintf("active PathHandler '%s' not found: %v", name, err),
			Code:    ErrPathNotFound.Code,
		}
	}

	return path, nil
}

// Helper method for pagination (reused from original implementation)
func (m *corePathCRUDManager) paginate(items *[]*conf.Path, itemsPerPage, page int) (int, error) {
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

// Helper method for API PathHandler pagination
func (m *corePathCRUDManager) paginateAPIPaths(items *[]*defs.APIPath, itemsPerPage, page int) (int, error) {
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

// paginatePathConf applies pagination to a slice of PathConf
func (m *corePathCRUDManager) paginatePathConf(items *[]PathConf, itemsPerPage, page int) (int, error) {
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
