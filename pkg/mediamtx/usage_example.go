package mediamtx

import (
	"fmt"
	"log"
)

// ExampleCorePathManagement demonstrates how to use the Core with path CRUD operations
func ExampleCorePathManagement() {
	// Create a MediaMTX Core instance
	core, err := New(Options{
		LogFunc: func(level LogLevel, format string, args ...interface{}) {
			log.Printf("[%v] %s", level, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		log.Fatalf("Failed to create Core: %v", err)
	}

	// Initialize the core
	if err := core.CreateResources(true); err != nil {
		log.Fatalf("Failed to create resources: %v", err)
	}

	// Example 1: Create different types of paths
	fmt.Println("Creating paths...")

	// Create a publisher path (accepts streams from publishers)
	if err := core.CreatePublisherPath("live_stream", true); err != nil {
		log.Printf("Failed to create publisher path: %v", err)
	} else {
		fmt.Println("✅ Created publisher path: live_stream")
	}

	// Create an RTSP path (pulls from RTSP camera)
	if err := core.CreateRTSPPath("camera1", "rtsp://camera.example.com/stream", false); err != nil {
		log.Printf("Failed to create RTSP path: %v", err)
	} else {
		fmt.Println("✅ Created RTSP path: camera1")
	}

	// Create a simple path using JSON
	jsonConfig := `{"source": "publisher", "record": false, "maxReaders": 50}`
	if err := core.CreatePathFromJSON("api_stream", jsonConfig); err != nil {
		log.Printf("Failed to create path from JSON: %v", err)
	} else {
		fmt.Println("✅ Created path from JSON: api_stream")
	}

	// Example 2: Manage existing paths
	fmt.Println("\nManaging paths...")

	// Check if paths exist
	paths := []string{"live_stream", "camera1", "api_stream", "nonexistent"}
	for _, pathName := range paths {
		if core.PathExists(pathName) {
			fmt.Printf("✅ Path '%s' exists\n", pathName)
		} else {
			fmt.Printf("❌ Path '%s' does not exist\n", pathName)
		}
	}

	// Get path information
	if pathInfo, err := core.GetPathInfo("live_stream"); err != nil {
		log.Printf("Failed to get path info: %v", err)
	} else {
		fmt.Printf("📊 Path 'live_stream' info: active=%v, type=%s\n", 
			pathInfo["active"], pathInfo["analysis"].(map[string]interface{})["type"])
	}

	// Example 3: Update paths
	fmt.Println("\nUpdating paths...")

	// Enable recording for a path
	if err := core.EnablePathRecording("camera1", "/recordings/camera1"); err != nil {
		log.Printf("Failed to enable recording: %v", err)
	} else {
		fmt.Println("✅ Enabled recording for camera1")
	}

	// Update path source
	if err := core.UpdatePathSource("camera1", "rtsp://new-camera.example.com/stream"); err != nil {
		log.Printf("Failed to update source: %v", err)
	} else {
		fmt.Println("✅ Updated camera1 source")
	}

	// Clone a path
	if err := core.ClonePath("live_stream", "live_stream_backup"); err != nil {
		log.Printf("Failed to clone path: %v", err)
	} else {
		fmt.Println("✅ Cloned live_stream to live_stream_backup")
	}

	// Example 4: Get statistics and overview
	fmt.Println("\nGetting statistics...")

	// Get path statistics
	if stats, err := core.GetPathStats(); err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		fmt.Printf("📊 Total paths: %v, Recording enabled: %v\n", 
			stats["total_configured"], stats["recording_enabled"])
		
		if byType, ok := stats["by_type"].(map[string]int); ok {
			fmt.Printf("📊 By type: %+v\n", byType)
		}
	}

	// Get comprehensive listing
	if allPaths, err := core.ListAllPaths(); err != nil {
		log.Printf("Failed to list all paths: %v", err)
	} else {
		fmt.Printf("📋 All paths: %d configured, %d active\n", 
			allPaths["total_configured"], allPaths["total_active"])
	}

	// Example 5: Using the PathCRUDManager directly
	fmt.Println("\nUsing PathCRUDManager directly...")

	pathManager := core.GetPathCRUDManager()

	// List paths with pagination
	if pathList, err := pathManager.ListPaths(2, 1); err != nil {
		log.Printf("Failed to list paths: %v", err)
	} else {
		fmt.Printf("📋 Page 1 (2 items): %d total, %d on this page\n", 
			pathList.ItemCount, len(pathList.Items))
	}

	// Validate a path configuration without creating it
	jsonConfig = `{"source": "rtsp://test.example.com/stream", "record": true}`
	if err := core.UpdatePathFromJSON("validation_test", jsonConfig); err != nil {
		fmt.Printf("✅ Validation correctly failed for non-existent path: %v\n", err)
	}

	// Example 6: Clean up
	fmt.Println("\nCleaning up...")

	pathsToRemove := []string{"api_stream", "live_stream_backup", "camera1"}
	for _, pathName := range pathsToRemove {
		if err := core.RemovePath(pathName); err != nil {
			log.Printf("Failed to remove path '%s': %v", pathName, err)
		} else {
			fmt.Printf("🗑️ Removed path: %s\n", pathName)
		}
	}

	fmt.Println("\n✅ Example completed successfully!")
}

// ExampleAdvancedPathOperations demonstrates advanced path operations
func ExampleAdvancedPathOperations() {
	// Create core instance (shortened for example)
	core, _ := New(Options{})
	_ = core.CreateResources(true)

	fmt.Println("Advanced Path Operations Example...")

	// Using helpers directly
	helper := NewPathHelper()
	validator := NewPathValidator()
	analyzer := NewPathAnalyzer()
	query := NewPathQuery()

	// Validate different path names
	pathNames := []string{"valid_path", "camera/1", "stream-2", "invalid name"}
	fmt.Println("\n🔍 Validating path names:")
	for _, name := range pathNames {
		if err := validator.ValidatePathName(name); err != nil {
			fmt.Printf("❌ '%s': %v\n", name, err)
		} else {
			fmt.Printf("✅ '%s' is valid\n", name)
		}
	}

	// Validate URLs
	urls := []string{
		"rtsp://camera.example.com/stream",
		"http://example.com/playlist.m3u8",
		"rtmp://server.example.com/app/stream",
	}
	fmt.Println("\n🔍 Validating URLs:")
	for _, url := range urls {
		if err := validator.ValidateRTSPURL(url); err == nil {
			fmt.Printf("✅ Valid RTSP URL: %s\n", url)
		}
		if err := validator.ValidateHLSURL(url); err == nil {
			fmt.Printf("✅ Valid HLS URL: %s\n", url)
		}
		if err := validator.ValidateRTMPURL(url); err == nil {
			fmt.Printf("✅ Valid RTMP URL: %s\n", url)
		}
	}

	// Create some test paths
	_ = core.CreatePublisherPath("live1", true)
	_ = core.CreateRTSPPath("camera1", "rtsp://cam1.example.com/stream", false)
	_ = core.CreateRTSPPath("camera2", "rtsp://cam2.example.com/stream", true)

	// Get all paths and analyze them
	if pathList, err := core.GetPathCRUDManager().ListPaths(0, 0); err == nil {
		fmt.Printf("\n📊 Analyzing %d paths:\n", len(pathList.Items))
		
		// Convert PathConf slice to conf.Path slice for compatibility
		confPaths, err := ConvertPathConfSliceToConfPaths(pathList.Items)
		if err != nil {
			fmt.Printf("❌ Failed to convert path configurations: %v\n", err)
		} else {
			for i, path := range confPaths {
				analysis := analyzer.AnalyzePathSource(path)
				fmt.Printf("   %s: type=%s, recording=%v\n", 
					pathList.Items[i].Name, analysis["type"], analysis["recording_enabled"])
			}

			// Filter paths by type
			rtspPaths := query.FilterPathsBySource(confPaths, "rtsp")
			fmt.Printf("\n🎥 RTSP paths: %d\n", len(rtspPaths))

			// Filter by recording
			recordingPaths := query.FilterPathsByRecording(confPaths, true)
			fmt.Printf("📹 Recording enabled paths: %d\n", len(recordingPaths))
		}
	}

	// Create path configurations using helpers
	fmt.Println("\n🛠️ Creating paths with helpers:")
	
	basicPath := helper.CreateBasicPath()
	rtspPath := helper.CreateRTSPPath("rtsp://example.com/stream")
	recordingPath := helper.CreateRecordingPath("/recordings/test", 0) // FMP4
	
	if basicPath != nil && rtspPath != nil && recordingPath != nil {
		fmt.Println("✅ All helper path configurations created successfully")
	}

	fmt.Println("\n✅ Advanced operations example completed!")
}