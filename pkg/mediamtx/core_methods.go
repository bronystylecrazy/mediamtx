package mediamtx

import (
	"encoding/json"
	"fmt"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/defs"
)

// Core-level convenience methods for path management

// createPathFromBuilder is a helper that converts builder to OptionalPath and creates the path
func (p *Core) createPathFromBuilder(name string, builder *PathConfigBuilder) error {
	pathConfig := builder.Build()
	optPath, err := pathConfig.ToOptionalPath()
	if err != nil {
		return fmt.Errorf("failed to convert path config: %v", err)
	}
	
	pathManager := p.GetPathCRUDManager()
	return pathManager.CreatePath(name, optPath)
}

// updatePathFromBuilder is a helper that converts builder to OptionalPath and updates the path
func (p *Core) updatePathFromBuilder(name string, builder *PathConfigBuilder) error {
	pathConfig := builder.Build()
	optPath, err := pathConfig.ToOptionalPath()
	if err != nil {
		return fmt.Errorf("failed to convert path config: %v", err)
	}
	
	pathManager := p.GetPathCRUDManager()
	return pathManager.UpdatePath(name, optPath)
}

// CreateSimplePath creates a path with simple parameters
func (p *Core) CreateSimplePath(name, source string, enableRecording bool) error {
	builder := NewSimplePathConfig(name, source, enableRecording)
	return p.createPathFromBuilder(name, builder)
}

// CreateRTSPPath creates a path configured for RTSP source
func (p *Core) CreateRTSPPath(name, rtspURL string, enableRecording bool) error {
	builder := NewRTSPPathConfig(name, rtspURL, enableRecording)
	return p.createPathFromBuilder(name, builder)
}

// CreateRTMPPath creates a path configured for RTMP source
func (p *Core) CreateRTMPPath(name, rtmpURL string, enableRecording bool) error {
	builder := NewSimplePathConfig(name, rtmpURL, enableRecording).SetSourceOnDemand(false)
	return p.createPathFromBuilder(name, builder)
}

// CreatePublisherPath creates a path that accepts publishers
func (p *Core) CreatePublisherPath(name string, enableRecording bool) error {
	builder := NewPublisherPathConfig(name, enableRecording)
	return p.createPathFromBuilder(name, builder)
}

// CreateOnDemandPath creates a path with on-demand activation
func (p *Core) CreateOnDemandPath(name, source, command string) error {
	builder := NewOnDemandPathConfig(name, source, command)
	return p.createPathFromBuilder(name, builder)
}

// UpdatePathSource updates the source of an existing path
func (p *Core) UpdatePathSource(name, newSource string) error {
	builder := NewPathConfigBuilder().SetSource(newSource)
	return p.updatePathFromBuilder(name, builder)
}

// EnablePathRecording enables recording for a path
func (p *Core) EnablePathRecording(name, recordPath string) error {
	builder := NewPathConfigBuilder().
		SetRecording(true).
		SetRecordPath(recordPath).
		SetRecordFormat("fmp4")
	return p.updatePathFromBuilder(name, builder)
}

// DisablePathRecording disables recording for a path
func (p *Core) DisablePathRecording(name string) error {
	builder := NewPathConfigBuilder().SetRecording(false)
	return p.updatePathFromBuilder(name, builder)
}

// GetPathInfo returns detailed information about a path (both config and runtime)
func (p *Core) GetPathInfo(name string) (map[string]interface{}, error) {
	pathManager := p.GetPathCRUDManager()
	
	// Get configuration
	pathConf, err := pathManager.GetPath(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get path configuration: %v", err)
	}

	info := make(map[string]interface{})
	info["configuration"] = pathConf

	// Try to get runtime information
	if apiPath, err := pathManager.GetActivePathInfo(name); err == nil {
		info["runtime"] = apiPath
		info["active"] = true
	} else {
		info["active"] = false
		info["runtime"] = nil
	}

	// Analyze the path
	analyzer := NewPathAnalyzer()
	analysis := analyzer.AnalyzePathSource(pathConf)
	info["analysis"] = analysis

	return info, nil
}

// ListAllPaths returns a comprehensive list of all paths with their information
func (p *Core) ListAllPaths() (map[string]interface{}, error) {
	pathManager := p.GetPathCRUDManager()
	
	// Get all configured paths
	pathList, err := pathManager.ListPaths(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list paths: %v", err)
	}

	// Get active paths
	activePaths, err := pathManager.GetActivePathsInfo(0, 0)
	if err != nil {
		// Log the error but don't fail - active paths might not be available
		p.Log(Warn, "Failed to get active paths info: %v", err)
		activePaths = &defs.APIPathList{Items: []*defs.APIPath{}}
	}

	// Create active paths map for quick lookup
	activePathsMap := make(map[string]*defs.APIPath)
	for _, activePath := range activePaths.Items {
		activePathsMap[activePath.Name] = activePath
	}

	// Combine information
	result := make(map[string]interface{})
	result["total_configured"] = pathList.ItemCount
	result["total_active"] = len(activePaths.Items)

	paths := make([]map[string]interface{}, len(pathList.Items))
	analyzer := NewPathAnalyzer()
	
	for i, pathConf := range pathList.Items {
		pathInfo := make(map[string]interface{})
		pathInfo["name"] = pathConf.Name
		pathInfo["configuration"] = pathConf
		
		// Add analysis - convert PathConf to conf.Path for analyzer
		confPath, err := pathConf.ToConfPath()
		if err != nil {
			// If conversion fails, provide basic analysis
			analysis := map[string]interface{}{
				"type": "unknown",
				"error": fmt.Sprintf("conversion failed: %v", err),
			}
			pathInfo["analysis"] = analysis
		} else {
			analysis := analyzer.AnalyzePathSource(confPath)
			pathInfo["analysis"] = analysis
		}
		
		// Add runtime info if available
		if activePath, exists := activePathsMap[pathConf.Name]; exists {
			pathInfo["runtime"] = activePath
			pathInfo["active"] = true
		} else {
			pathInfo["active"] = false
			pathInfo["runtime"] = nil
		}
		
		paths[i] = pathInfo
	}
	
	result["paths"] = paths
	return result, nil
}

// GetPathStats returns statistics about all paths
func (p *Core) GetPathStats() (map[string]interface{}, error) {
	pathManager := p.GetPathCRUDManager()
	
	// Get configured paths
	pathList, err := pathManager.ListPaths(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list paths: %v", err)
	}

	// Get active paths
	activePaths, err := pathManager.GetActivePathsInfo(0, 0)
	if err != nil {
		p.Log(Warn, "Failed to get active paths info: %v", err)
		activePaths = &defs.APIPathList{Items: []*defs.APIPath{}}
	}

	stats := NewPathStats()
	query := NewPathQuery()

	result := make(map[string]interface{})
	
	// Basic counts
	result["total_configured"] = pathList.ItemCount
	result["total_active"] = len(activePaths.Items)
	
	// Convert PathConf slice to conf.Path slice for compatibility
	confPaths, err := ConvertPathConfSliceToConfPaths(pathList.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path configurations: %v", err)
	}
	
	// Count by type
	result["by_type"] = stats.CountPathsByType(confPaths)
	
	// Count by features
	recordingPaths := query.FilterPathsByRecording(confPaths, true)
	result["recording_enabled"] = len(recordingPaths)
	
	onDemandPaths := query.FilterPathsByOnDemand(confPaths, true)
	result["on_demand_enabled"] = len(onDemandPaths)
	
	// Traffic statistics from active paths
	if len(activePaths.Items) > 0 {
		trafficStats := stats.CalculateTrafficStats(activePaths.Items)
		result["traffic"] = trafficStats
	}
	
	return result, nil
}

// RemovePath removes a path (alias for DeletePath for convenience)
func (p *Core) RemovePath(name string) error {
	pathManager := p.GetPathCRUDManager()
	return pathManager.DeletePath(name)
}

// PathExists checks if a path exists
func (p *Core) PathExists(name string) bool {
	pathManager := p.GetPathCRUDManager()
	_, err := pathManager.GetPath(name)
	return err == nil
}

// IsPathActive checks if a path is currently active
func (p *Core) IsPathActive(name string) bool {
	pathManager := p.GetPathCRUDManager()
	_, err := pathManager.GetActivePathInfo(name)
	return err == nil
}

// ClonePath creates a copy of an existing path with a new name
func (p *Core) ClonePath(sourceName, targetName string) error {
	pathManager := p.GetPathCRUDManager()
	
	// Get the source path and convert to OptionalPath directly
	sourcePath, err := pathManager.GetPath(sourceName)
	if err != nil {
		return fmt.Errorf("failed to get source path '%s': %v", sourceName, err)
	}

	// Convert to JSON and back to OptionalPath for cloning
	jsonData, err := json.Marshal(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to marshal source path: %v", err)
	}

	var optPath conf.OptionalPath
	if err := json.Unmarshal(jsonData, &optPath); err != nil {
		return fmt.Errorf("failed to unmarshal path config: %v", err)
	}

	// Create the cloned path
	return pathManager.CreatePath(targetName, &optPath)
}

// pathFromJSON is a helper that parses JSON config to OptionalPath
func (p *Core) pathFromJSON(jsonConfig string) (*conf.OptionalPath, error) {
	var optPath conf.OptionalPath
	if err := json.Unmarshal([]byte(jsonConfig), &optPath); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON config: %v", err)
	}
	return &optPath, nil
}

// UpdatePathFromJSON updates a path using JSON configuration
func (p *Core) UpdatePathFromJSON(name string, jsonConfig string) error {
	optPath, err := p.pathFromJSON(jsonConfig)
	if err != nil {
		return err
	}
	return p.GetPathCRUDManager().UpdatePath(name, optPath)
}

// CreatePathFromJSON creates a path using JSON configuration
func (p *Core) CreatePathFromJSON(name string, jsonConfig string) error {
	optPath, err := p.pathFromJSON(jsonConfig)
	if err != nil {
		return err
	}
	return p.GetPathCRUDManager().CreatePath(name, optPath)
}