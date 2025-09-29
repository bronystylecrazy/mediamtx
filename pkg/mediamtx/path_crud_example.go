package mediamtx

import (
	"encoding/json"
	"fmt"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/defs"
)

// Example usage of the PathCRUDManager

// ExamplePathCRUDUsage demonstrates how to use the PathCRUDManager
func ExamplePathCRUDUsage() {
	// Create a basic configuration
	config := &conf.Conf{
		Paths: make(map[string]*conf.Path),
	}

	// Create PathHandler CRUD manager
	pathManager := NewPathCRUDManager(config, nil, func(newConf *conf.Conf) {
		fmt.Printf("Configuration updated with %d paths\n", len(newConf.Paths))
	})

	// Create a simple PathHandler using JSON
	pathJSON := `{"source": "publisher", "record": false}`
	var optPath conf.OptionalPath
	if err := json.Unmarshal([]byte(pathJSON), &optPath); err != nil {
		fmt.Printf("Error creating PathHandler: %v\n", err)
		return
	}

	// Create the PathHandler
	if err := pathManager.CreatePath("test_stream", &optPath); err != nil {
		fmt.Printf("Error creating PathHandler: %v\n", err)
	} else {
		fmt.Println("✅ Path created successfully")
	}

	// List paths
	if pathList, err := pathManager.ListPaths(0, 0); err != nil {
		fmt.Printf("Error listing paths: %v\n", err)
	} else {
		fmt.Printf("Found %d paths\n", pathList.ItemCount)
	}

	// Get PathHandler
	if path, err := pathManager.GetPath("test_stream"); err != nil {
		fmt.Printf("Error getting PathHandler: %v\n", err)
	} else {
		fmt.Printf("Path source: %s\n", path.Source)
	}

	// Update PathHandler to enable recording
	updateJSON := `{"record": true, "recordPath": "/recordings"}`
	var updatePath conf.OptionalPath
	if err := json.Unmarshal([]byte(updateJSON), &updatePath); err != nil {
		fmt.Printf("Error parsing update: %v\n", err)
		return
	}

	if err := pathManager.UpdatePath("test_stream", &updatePath); err != nil {
		fmt.Printf("Error updating PathHandler: %v\n", err)
	} else {
		fmt.Println("✅ Path updated successfully")
	}

	// Delete PathHandler
	if err := pathManager.DeletePath("test_stream"); err != nil {
		fmt.Printf("Error deleting PathHandler: %v\n", err)
	} else {
		fmt.Println("✅ Path deleted successfully")
	}
}

// ExamplePathHelperUsage demonstrates helper utilities
func ExamplePathHelperUsage() {
	_ = NewPathHelper() // Keep for other functionality
	validator := NewPathValidator()
	analyzer := NewPathAnalyzer()

	// Validate PathHandler names
	testNames := []string{"valid_path", "camera1", "stream/sub", "invalid name"}
	for _, name := range testNames {
		if err := validator.ValidatePathName(name); err != nil {
			fmt.Printf("❌ '%s': %v\n", name, err)
		} else {
			fmt.Printf("✅ '%s' is valid\n", name)
		}
	}

	// Validate URLs
	urls := []string{
		"rtsp://camera.example.com/stream",
		"http://example.com/stream.m3u8",
		"invalid://url",
	}

	for _, url := range urls {
		if err := validator.ValidateRTSPURL(url); err == nil {
			fmt.Printf("✅ RTSP URL '%s' is valid\n", url)
		}
		if err := validator.ValidateHLSURL(url); err == nil {
			fmt.Printf("✅ HLS URL '%s' is valid\n", url)
		}
	}

	// Create PathHandler configurations using new factory methods
	basicPath := NewSimplePathConfig("example_basic", "publisher", false).Build()
	rtspPath := NewRTSPPathConfig("example_rtsp", "rtsp://camera.example.com/stream", false).Build()

	fmt.Printf("Created basic PathHandler: %v\n", basicPath.Name != "")
	fmt.Printf("Created RTSP PathHandler: %v\n", rtspPath.Name != "")

	// Analyze paths (requires actual Path objects, not OptionalPath)
	samplePath := &conf.Path{
		Source: "rtsp://camera.example.com/stream",
		Record: true,
	}

	analysis := analyzer.AnalyzePathSource(samplePath)
	fmt.Printf("Path analysis: type=%s, description=%s\n",
		analysis["type"], analysis["description"])
}

// ExampleAPIPathInfo demonstrates working with active PathHandler information
func ExampleAPIPathInfo() {
	// This would typically be called with an actual PathHandler manager that has access to active paths
	// For demonstration, we'll create mock data

	mockAPIPath := &defs.APIPath{
		Name:          "camera1",
		ConfName:      "camera1",
		Ready:         true,
		Tracks:        []string{"video", "audio"},
		BytesReceived: 1024000,
		BytesSent:     512000,
		Source: &defs.APIPathSourceOrReader{
			Type: "rtsp",
			ID:   "source-1",
		},
		Readers: []defs.APIPathSourceOrReader{
			{Type: "hls", ID: "reader-1"},
			{Type: "webrtc", ID: "reader-2"},
		},
	}

	analyzer := NewPathAnalyzer()
	status := analyzer.AnalyzePathStatus(mockAPIPath)

	fmt.Printf("Path Status Analysis:\n")
	fmt.Printf("  Name: %s\n", status["name"])
	fmt.Printf("  Ready: %v\n", status["ready"])
	fmt.Printf("  Tracks: %d\n", status["track_count"])
	fmt.Printf("  Bytes Received: %v\n", status["bytes_received"])
	fmt.Printf("  Bytes Sent: %v\n", status["bytes_sent"])
	fmt.Printf("  Reader Count: %d\n", status["reader_count"])
}

// PathCRUDManagerAPI provides a high-level API wrapper for common operations
type PathCRUDManagerAPI struct {
	manager PathCRUDManager
	helper  *PathHelper
}

// NewPathCRUDManagerAPI creates a new high-level API wrapper
func NewPathCRUDManagerAPI(manager PathCRUDManager) *PathCRUDManagerAPI {
	return &PathCRUDManagerAPI{
		manager: manager,
		helper:  NewPathHelper(),
	}
}

// CreateSimplePath creates a PathHandler with simple JSON configuration
func (api *PathCRUDManagerAPI) CreateSimplePath(name, source string, enableRecording bool) error {
	pathConfig := map[string]interface{}{
		"source": source,
		"record": enableRecording,
	}

	if enableRecording {
		pathConfig["recordPath"] = fmt.Sprintf("/recordings/%s", name)
		pathConfig["recordFormat"] = 0 // FMP4
	}

	jsonData, err := json.Marshal(pathConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal PathHandler config: %v", err)
	}

	var optPath conf.OptionalPath
	if err := json.Unmarshal(jsonData, &optPath); err != nil {
		return fmt.Errorf("failed to unmarshal PathHandler config: %v", err)
	}

	return api.manager.CreatePath(name, &optPath)
}

// EnableRecording enables recording for an existing PathHandler
func (api *PathCRUDManagerAPI) EnableRecording(name, recordPath string) error {
	updateConfig := map[string]interface{}{
		"record":     true,
		"recordPath": recordPath,
	}

	jsonData, err := json.Marshal(updateConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal update config: %v", err)
	}

	var updatePath conf.OptionalPath
	if err := json.Unmarshal(jsonData, &updatePath); err != nil {
		return fmt.Errorf("failed to unmarshal update config: %v", err)
	}

	return api.manager.UpdatePath(name, &updatePath)
}

// GetPathSummary returns a summary of all paths
func (api *PathCRUDManagerAPI) GetPathSummary() (map[string]interface{}, error) {
	pathList, err := api.manager.ListPaths(0, 0)
	if err != nil {
		return nil, err
	}

	query := NewPathQuery()
	stats := NewPathStats()

	summary := make(map[string]interface{})
	summary["total_paths"] = pathList.ItemCount

	// Convert PathConf slice to conf.Path slice for compatibility
	confPaths, err := ConvertPathConfSliceToConfPaths(pathList.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PathHandler configurations: %v", err)
	}

	// Count by type
	typeCounts := stats.CountPathsByType(confPaths)
	summary["by_type"] = typeCounts

	// Count recording enabled
	recordingPaths := query.FilterPathsByRecording(confPaths, true)
	summary["recording_enabled"] = len(recordingPaths)

	// Count on-demand
	onDemandPaths := query.FilterPathsByOnDemand(confPaths, true)
	summary["on_demand_enabled"] = len(onDemandPaths)

	return summary, nil
}
