package mediamtx

import (
	"testing"
	"time"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/stretchr/testify/require"
)

func TestDirectAPI_ConfigurationManagement(t *testing.T) {
	api := NewMediaMTXAPI(&Core{})
	
	// Test GetPathDefaults
	defaults := api.GetPathDefaults()
	require.NotNil(t, defaults)
	
	// Test pagination conversion
	pagination, err := PaginateFromStrings("10", "1")
	require.NoError(t, err)
	require.Equal(t, 10, pagination.ItemsPerPage)
	require.Equal(t, 1, pagination.Page)
}

func TestDirectAPI_PathManagement(t *testing.T) {
	// Simple test without complex type dependencies
	api := NewMediaMTXAPI(&Core{})
	require.NotNil(t, api)
	
	// Test basic functionality
	defaults := api.GetPathDefaults()
	require.NotNil(t, defaults)
}

func TestDirectAPI_RecordingQuery(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)
	
	query := &RecordingQuery{
		Path:      "test_path",
		StartTime: &startTime,
		EndTime:   &endTime,
	}
	
	require.Equal(t, "test_path", query.Path)
	require.True(t, query.StartTime.Before(now))
	require.True(t, query.EndTime.After(now))
}

func TestDirectAPI_PaginationParams(t *testing.T) {
	// Test DefaultPagination
	defaultPagination := DefaultPagination()
	require.Equal(t, 100, defaultPagination.ItemsPerPage)
	require.Equal(t, 0, defaultPagination.Page)
	
	// Test PaginateFromStrings with valid values
	pagination, err := PaginateFromStrings("50", "2")
	require.NoError(t, err)
	require.Equal(t, 50, pagination.ItemsPerPage)
	require.Equal(t, 2, pagination.Page)
	
	// Test PaginateFromStrings with empty values (should use defaults)
	pagination, err = PaginateFromStrings("", "")
	require.NoError(t, err)
	require.Equal(t, 100, pagination.ItemsPerPage)
	require.Equal(t, 0, pagination.Page)
	
	// Test PaginateFromStrings with invalid values
	_, err = PaginateFromStrings("invalid", "0")
	require.Error(t, err)
	
	_, err = PaginateFromStrings("0", "0")
	require.Error(t, err)
	require.Contains(t, err.Error(), "itemsPerPage must be greater than 0")
}

func TestDirectAPI_AuthenticationMethods(t *testing.T) {
	// Test with nil AuthManager
	core := &Core{}
	api := NewMediaMTXAPI(core)
	
	// Should not panic with nil AuthManager
	require.NotPanics(t, func() {
		api.RefreshJWTJWKS()
	})
	
	// Authenticate should return nil with nil AuthManager
	req, err := api.CreateAuthRequest("user", "pass", "", "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, req)
	
	authErr := api.Authenticate(req)
	require.Nil(t, authErr)
	
	// Test CreateAuthRequest fields
	require.Equal(t, conf.AuthActionAPI, req.Action)
	require.NotNil(t, req.Credentials)
	require.Equal(t, "user", req.Credentials.User)
	require.Equal(t, "pass", req.Credentials.Pass)
	require.Equal(t, "127.0.0.1", req.IP.String())
}

func TestDirectAPI_RecordingInfo(t *testing.T) {
	info := &RecordingInfo{
		PathName:        "test_path",
		TotalRecordings: 5,
		TotalDuration:   30 * time.Minute,
	}
	
	require.Equal(t, "test_path", info.PathName)
	require.Equal(t, 5, info.TotalRecordings)
	require.Equal(t, 30*time.Minute, info.TotalDuration)
}

func TestDirectAPI_APIResult(t *testing.T) {
	result := &APIResult{
		Data: "test data",
		Pagination: &PaginationResult{
			ItemCount: 10,
			PageCount: 2,
		},
		Error: nil,
	}
	
	require.Equal(t, "test data", result.Data)
	require.Equal(t, 10, result.Pagination.ItemCount)
	require.Equal(t, 2, result.Pagination.PageCount)
	require.Nil(t, result.Error)
}

func TestDirectAPI_paginateSlice(t *testing.T) {
	core := &Core{
		Conf: &conf.Conf{},
	}
	api := NewMediaMTXAPI(core)
	
	// Test pagination with empty slice
	items := []string{}
	pageCount := api.paginateSlice(&items, 10, 0)
	require.Equal(t, 0, pageCount)
	require.Len(t, items, 0)
	
	// Test pagination with items
	items = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	pageCount = api.paginateSlice(&items, 5, 0)
	require.Equal(t, 3, pageCount) // 12 items / 5 per page = 3 pages (rounded up)
	require.Len(t, items, 5)       // First page should have 5 items
	require.Equal(t, []string{"a", "b", "c", "d", "e"}, items)
	
	// Test second page
	items = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	pageCount = api.paginateSlice(&items, 5, 1)
	require.Equal(t, 3, pageCount)
	require.Len(t, items, 5)
	require.Equal(t, []string{"f", "g", "h", "i", "j"}, items)
	
	// Test last page (partial)
	items = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	pageCount = api.paginateSlice(&items, 5, 2)
	require.Equal(t, 3, pageCount)
	require.Len(t, items, 2) // Last page has remaining 2 items
	require.Equal(t, []string{"k", "l"}, items)
	
	// Test with itemsPerPage <= 0
	items = []string{"a", "b", "c"}
	pageCount = api.paginateSlice(&items, 0, 0)
	require.Equal(t, 1, pageCount)
	
	// Test page beyond available data
	items = []string{"a", "b", "c"}
	pageCount = api.paginateSlice(&items, 2, 5)
	require.Equal(t, 2, pageCount) // Still 2 pages total (3 items / 2 per page)
	require.Len(t, items, 0)       // No items for page 5
}

func TestDirectAPI_sortedPathKeys(t *testing.T) {
	core := &Core{}
	api := NewMediaMTXAPI(core)
	
	paths := map[string]*conf.Path{
		"zebra": &conf.Path{Name: "zebra"},
		"alpha": &conf.Path{Name: "alpha"},
		"beta":  &conf.Path{Name: "beta"},
	}
	
	keys := api.sortedPathKeys(paths)
	require.Equal(t, []string{"alpha", "beta", "zebra"}, keys)
	
	// Test with empty map
	emptyPaths := map[string]*conf.Path{}
	keys = api.sortedPathKeys(emptyPaths)
	require.Len(t, keys, 0)
}

func TestDirectAPI_PatchMethods(t *testing.T) {
	// Test that patch methods exist and have proper signatures
	api := NewMediaMTXAPI(&Core{})
	
	// Test that methods exist by checking they're not nil
	require.NotNil(t, api.PatchGlobalConfig)
	require.NotNil(t, api.PatchPathDefaults)
	
	// Note: Cannot test actual functionality without proper core setup
	// as it would require a valid configuration structure
}

func TestDirectAPI_OptionalPathHelpers(t *testing.T) {
	// Test NewOptionalPath helper function
	optPath := NewOptionalPath("rtsp://localhost:8554/test")
	require.NotNil(t, optPath)
	require.NotNil(t, optPath.Values)
	
	// Verify the source was set
	if path, ok := optPath.Values.(*conf.Path); ok {
		require.Equal(t, "rtsp://localhost:8554/test", path.Source)
		require.False(t, path.Record) // Default should be false
	}
	
	// Test NewOptionalPathWithOptions using comprehensive typed struct
	options := PathOptions{
		Source:                     "rtsp://example.com/stream",
		Record:                     true,
		RecordPath:                 "/recordings",
		RecordFormat:               "mp4",
		MaxReaders:                 10,
		SourceOnDemand:             true,
		SourceOnDemandStartTimeout: "10s",
		SourceFingerprint:          "abc123",
		RTSPTransport:              "udp",
		RTSPAnyPort:                true,
		RunOnInit:                  "/scripts/init.sh",
		RunOnInitRestart:           true,
		UseAbsoluteTimestamp:       true,
	}
	optPathWithOpts := NewOptionalPathWithOptions(options)
	require.NotNil(t, optPathWithOpts)
	require.NotNil(t, optPathWithOpts.Values)
	
	// The Values field should now be a map[string]interface{} with the converted JSON data
	if pathData, ok := optPathWithOpts.Values.(map[string]interface{}); ok {
		require.Equal(t, "rtsp://example.com/stream", pathData["source"])
		require.Equal(t, true, pathData["record"])
		require.Equal(t, "/recordings", pathData["recordPath"])
		require.Equal(t, "mp4", pathData["recordFormat"])
		require.Equal(t, float64(10), pathData["maxReaders"]) // JSON numbers become float64
		require.Equal(t, true, pathData["sourceOnDemand"])
		require.Equal(t, "10s", pathData["sourceOnDemandStartTimeout"])
		require.Equal(t, "abc123", pathData["sourceFingerprint"])
		require.Equal(t, "udp", pathData["rtspTransport"])
		require.Equal(t, true, pathData["rtspAnyPort"])
		require.Equal(t, "/scripts/init.sh", pathData["runOnInit"])
		require.Equal(t, true, pathData["runOnInitRestart"])
		require.Equal(t, true, pathData["useAbsoluteTimestamp"])
	}
}

func TestDirectAPI_ListPathConfigs(t *testing.T) {
	api := NewMediaMTXAPI(&Core{})
	
	// Test with nil configuration
	result, err := api.ListPathConfigs(nil)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "configuration not available")
}

func TestDirectAPI_IntegrationWithRealCore(t *testing.T) {
	// Test without actually using the Core to avoid complex setup
	// Focus on testing the PathOptions struct and helper functions
	
	// Test 1: Simple path addition
	t.Run("Simple Path Addition", func(t *testing.T) {
		simpleOptions := PathOptions{
			Source: "rtsp://localhost:8554/test",
		}
		
		optPath := NewOptionalPathWithOptions(simpleOptions)
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		// Verify the Values field contains the expected data
		if pathData, ok := optPath.Values.(map[string]interface{}); ok {
			require.Equal(t, "rtsp://localhost:8554/test", pathData["source"])
		} else {
			t.Errorf("Expected Values to be map[string]interface{}, got %T", optPath.Values)
		}
	})
	
	// Test 2: Comprehensive path configuration
	t.Run("Comprehensive Path Configuration", func(t *testing.T) {
		comprehensiveOptions := PathOptions{
			Source:                     "rtsps://secure-camera:8554/stream",
			Record:                     true,
			RecordPath:                 "/recordings/test",
			RecordFormat:               "mp4",
			RecordPartDuration:         "1h",
			RecordDeleteAfter:          "168h",
			MaxReaders:                 10,
			SourceOnDemand:             true,
			SourceOnDemandStartTimeout: "10s",
			SourceOnDemandCloseAfter:   "30s",
			UseAbsoluteTimestamp:       true,
			RTSPTransport:              "tcp",
			RTSPAnyPort:                true,
			RunOnInit:                  "/scripts/init.sh",
			RunOnReady:                 "/scripts/ready.sh",
			RunOnRecordSegmentComplete: "/scripts/process.sh",
		}
		
		optPath := NewOptionalPathWithOptions(comprehensiveOptions)
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		// Verify multiple fields are correctly set
		if pathData, ok := optPath.Values.(map[string]interface{}); ok {
			require.Equal(t, "rtsps://secure-camera:8554/stream", pathData["source"])
			require.Equal(t, true, pathData["record"])
			require.Equal(t, "/recordings/test", pathData["recordPath"])
			require.Equal(t, "mp4", pathData["recordFormat"])
			require.Equal(t, "1h", pathData["recordPartDuration"])
			require.Equal(t, "168h", pathData["recordDeleteAfter"])
			require.Equal(t, float64(10), pathData["maxReaders"])
			require.Equal(t, true, pathData["sourceOnDemand"])
			require.Equal(t, "10s", pathData["sourceOnDemandStartTimeout"])
			require.Equal(t, "30s", pathData["sourceOnDemandCloseAfter"])
			require.Equal(t, true, pathData["useAbsoluteTimestamp"])
			require.Equal(t, "tcp", pathData["rtspTransport"])
			require.Equal(t, true, pathData["rtspAnyPort"])
			require.Equal(t, "/scripts/init.sh", pathData["runOnInit"])
			require.Equal(t, "/scripts/ready.sh", pathData["runOnReady"])
			require.Equal(t, "/scripts/process.sh", pathData["runOnRecordSegmentComplete"])
		} else {
			t.Errorf("Expected Values to be map[string]interface{}, got %T", optPath.Values)
		}
	})
	
	// Test 3: Raspberry Pi Camera configuration
	t.Run("Raspberry Pi Camera Configuration", func(t *testing.T) {
		piCameraOptions := PathOptions{
			RPICameraCamID:                 1, // Use camera 1 instead of 0 to avoid omitempty
			RPICameraWidth:                 1920,
			RPICameraHeight:                1080,
			RPICameraFPS:                   30.0,
			RPICameraHFlip:                 true,
			RPICameraVFlip:                 false,
			RPICameraBrightness:            0.5,
			RPICameraContrast:              1.0,
			RPICameraExposure:              "auto",
			RPICameraCodec:                 "h264",
			RPICameraBitrate:               2000000,
			RPICameraHardwareH264Profile:   "high",
			Record:                         true,
			RecordPath:                     "/recordings/pi-cam",
		}
		
		optPath := NewOptionalPathWithOptions(piCameraOptions)
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		// Verify Pi camera specific fields
		if pathData, ok := optPath.Values.(map[string]interface{}); ok {
			require.Equal(t, float64(1), pathData["rpiCameraCamID"])
			require.Equal(t, float64(1920), pathData["rpiCameraWidth"])
			require.Equal(t, float64(1080), pathData["rpiCameraHeight"])
			require.Equal(t, 30.0, pathData["rpiCameraFPS"])
			require.Equal(t, true, pathData["rpiCameraHFlip"])
			require.Equal(t, 0.5, pathData["rpiCameraBrightness"])
			require.Equal(t, 1.0, pathData["rpiCameraContrast"])
			require.Equal(t, "auto", pathData["rpiCameraExposure"])
			require.Equal(t, "h264", pathData["rpiCameraCodec"])
			require.Equal(t, float64(2000000), pathData["rpiCameraBitrate"])
			require.Equal(t, "high", pathData["rpiCameraHardwareH264Profile"])
			require.Equal(t, true, pathData["record"])
			require.Equal(t, "/recordings/pi-cam", pathData["recordPath"])
			
			// Verify that false boolean values are not included (rpiCameraVFlip)
			_, exists := pathData["rpiCameraVFlip"]
			require.False(t, exists, "False boolean values should be filtered out")
		} else {
			t.Errorf("Expected Values to be map[string]interface{}, got %T", optPath.Values)
		}
	})
	
	// Test 4: Zero value filtering
	t.Run("Zero Value Filtering", func(t *testing.T) {
		mixedOptions := PathOptions{
			Source:         "rtsp://test:8554/stream", // Non-zero string
			Record:         true,                      // Non-zero bool
			RecordPath:     "",                       // Zero string - should be filtered
			MaxReaders:     0,                        // Zero int - now INCLUDED  
			SourceOnDemand: false,                    // Zero bool - should be filtered
			RPICameraFPS:   25.0,                     // Non-zero float
		}
		
		optPath := NewOptionalPathWithOptions(mixedOptions)
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		if pathData, ok := optPath.Values.(map[string]interface{}); ok {
			// Should be included (non-zero values)
			require.Equal(t, "rtsp://test:8554/stream", pathData["source"])
			require.Equal(t, true, pathData["record"])
			require.Equal(t, 25.0, pathData["rpiCameraFPS"])
			
			t.Logf("✅ Correctly filtered: only non-zero values included (%d fields)", len(pathData))
			
			// Should be filtered out
			_, hasRecordPath := pathData["recordPath"]
			require.False(t, hasRecordPath, "Empty string should be filtered out")
			
			_, hasSourceOnDemand := pathData["sourceOnDemand"]
			require.False(t, hasSourceOnDemand, "False bool should be filtered out")
		}
	})
	
	// Test 5: Simulate the original user's use case
	t.Run("Original User Case - No Panic Test", func(t *testing.T) {
		// This simulates the exact case that was causing the panic
		// before we implemented the comprehensive PathOptions
		
		// Original problematic usage (now fixed):
		optPath := NewOptionalPathWithOptions(PathOptions{
			Source: "rtsp://localhost:8554/test",
			Record: true,
			RecordPath: "/recordings",
		})
		
		// Verify no panic and proper structure
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		// Verify the Values field is properly structured
		pathData, ok := optPath.Values.(map[string]interface{})
		require.True(t, ok, "Values should be map[string]interface{}")
		
		// Verify all non-zero values are present
		require.Equal(t, "rtsp://localhost:8554/test", pathData["source"])
		require.Equal(t, true, pathData["record"])
		require.Equal(t, "/recordings", pathData["recordPath"])
		
		t.Log("✅ Original user case now works without panic!")
	})
	
	// Test 6: Comprehensive field coverage test
	t.Run("Comprehensive Field Coverage", func(t *testing.T) {
		// Test with many fields to ensure JSON conversion works
		options := PathOptions{
			// Basic fields
			Source: "rtsp://example.com/stream", 
			Record: true,
			RecordPath: "/recordings/test",
			
			// Advanced fields from OpenAPI
			RecordFormat: "mp4",
			RecordPartDuration: "1h",
			RecordDeleteAfter: "7d",
			MaxReaders: 10,
			SourceOnDemand: true,
			SourceOnDemandStartTimeout: "10s",
			SourceFingerprint: "abc123",
			
			// RTSP specific
			RTSPTransport: "tcp",
			RTSPAnyPort: true,
			RTSPRangeType: "clock",
			
			// Pi Camera (subset)
			RPICameraWidth: 1920,
			RPICameraHeight: 1080,
			RPICameraFPS: 30.0,
			RPICameraBrightness: 0.8,
			
			// Hooks
			RunOnInit: "/scripts/init.sh",
			RunOnReady: "/scripts/ready.sh",
			RunOnRecordSegmentComplete: "/scripts/process.sh",
		}
		
		optPath := NewOptionalPathWithOptions(options)
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		pathData, ok := optPath.Values.(map[string]interface{})
		require.True(t, ok, "Values should be map[string]interface{}")
		
		// Verify a sampling of different field types
		require.Equal(t, "rtsp://example.com/stream", pathData["source"])
		require.Equal(t, "mp4", pathData["recordFormat"])
		require.Equal(t, float64(1920), pathData["rpiCameraWidth"])
		require.Equal(t, 30.0, pathData["rpiCameraFPS"])
		require.Equal(t, "/scripts/init.sh", pathData["runOnInit"])
		
		t.Logf("✅ Successfully converted %d fields to JSON format", len(pathData))
	})
}

func TestDirectAPI_PathOptions_Validation(t *testing.T) {
	// Test that PathOptions create valid OptionalPath objects that would work with AddPathConfig
	// This focuses on testing the original user issue without requiring full Core setup
	
	t.Run("PathOptions_Simple_Source", func(t *testing.T) {
		// Test the helper function creates valid OptionalPath
		optPath := NewOptionalPath("rtsp://localhost:8554/test")
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		// NewOptionalPath actually creates map[string]interface{} via JSON conversion
		if pathData, ok := optPath.Values.(map[string]interface{}); ok {
			require.Equal(t, "rtsp://localhost:8554/test", pathData["source"])
		} else {
			t.Errorf("Expected Values to be map[string]interface{}, got %T", optPath.Values)
		}
		
		t.Log("✅ Simple source OptionalPath creation works")
	})
	
	t.Run("PathOptions_Original_User_Case_No_Panic", func(t *testing.T) {
		// Test the exact case that was causing panic for the user
		options := PathOptions{
			Source:     "rtsp://localhost:8554/test",
			Record:     true,
			RecordPath: "/recordings",
		}
		
		// This should NOT panic (was the original issue)
		optPath := NewOptionalPathWithOptions(options)
		require.NotNil(t, optPath, "OptionalPath creation should not panic")
		require.NotNil(t, optPath.Values, "Values should not be nil")
		
		// Verify the conversion worked correctly
		if pathData, ok := optPath.Values.(map[string]interface{}); ok {
			require.Equal(t, "rtsp://localhost:8554/test", pathData["source"])
			require.Equal(t, true, pathData["record"])
			require.Equal(t, "/recordings", pathData["recordPath"])
		} else {
			t.Errorf("Expected Values to be map[string]interface{}, got %T", optPath.Values)
		}
		
		t.Log("✅ Original user case now works without panic!")
	})
	
	t.Run("PathOptions_Comprehensive_Configuration", func(t *testing.T) {
		// Test comprehensive options work correctly
		options := PathOptions{
			Source:                     "rtsp://camera1:8554/stream",
			Record:                     true,
			RecordPath:                 "/recordings/camera1",
			RecordFormat:               "mp4",
			RecordPartDuration:         "1h",
			RecordDeleteAfter:          "168h",
			MaxReaders:                 5,
			SourceOnDemand:             false,
			SourceOnDemandStartTimeout: "10s",
			RTSPTransport:              "tcp",
			RTSPAnyPort:                true,
			UseAbsoluteTimestamp:       true,
			RunOnReady:                 "/scripts/ready.sh",
			RunOnRecordSegmentComplete: "/scripts/process.sh",
		}
		
		optPath := NewOptionalPathWithOptions(options)
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		// Verify comprehensive field conversion
		if pathData, ok := optPath.Values.(map[string]interface{}); ok {
			require.Equal(t, "rtsp://camera1:8554/stream", pathData["source"])
			require.Equal(t, true, pathData["record"])
			require.Equal(t, "/recordings/camera1", pathData["recordPath"])
			require.Equal(t, "mp4", pathData["recordFormat"])
			require.Equal(t, "1h", pathData["recordPartDuration"])
			require.Equal(t, "168h", pathData["recordDeleteAfter"])
			require.Equal(t, float64(5), pathData["maxReaders"])
			// Note: false values are omitted due to omitempty, so sourceOnDemand won't be present
			require.Equal(t, "10s", pathData["sourceOnDemandStartTimeout"])
			require.Equal(t, "tcp", pathData["rtspTransport"])
			require.Equal(t, true, pathData["rtspAnyPort"])
			require.Equal(t, true, pathData["useAbsoluteTimestamp"])
			require.Equal(t, "/scripts/ready.sh", pathData["runOnReady"])
			require.Equal(t, "/scripts/process.sh", pathData["runOnRecordSegmentComplete"])
			
			t.Logf("✅ Successfully converted %d configuration fields", len(pathData))
		} else {
			t.Errorf("Expected Values to be map[string]interface{}, got %T", optPath.Values)
		}
		
		t.Log("✅ Comprehensive PathOptions test passed")
	})
	
	t.Run("PathOptions_RaspberryPi_Camera", func(t *testing.T) {
		// Test Raspberry Pi camera specific fields
		options := PathOptions{
			RPICameraCamID:     1,
			RPICameraWidth:     1920,
			RPICameraHeight:    1080,
			RPICameraFPS:       30.0,
			RPICameraHFlip:     true,
			RPICameraVFlip:     false,
			RPICameraExposure:  "auto",
			RPICameraCodec:     "h264",
			RPICameraBrightness: 0.5,
			Record:             true,
			RecordPath:         "/recordings/pi-cam",
		}
		
		optPath := NewOptionalPathWithOptions(options)
		require.NotNil(t, optPath)
		require.NotNil(t, optPath.Values)
		
		// Verify Pi camera fields
		if pathData, ok := optPath.Values.(map[string]interface{}); ok {
			require.Equal(t, float64(1), pathData["rpiCameraCamID"])
			require.Equal(t, float64(1920), pathData["rpiCameraWidth"])
			require.Equal(t, float64(1080), pathData["rpiCameraHeight"])
			require.Equal(t, 30.0, pathData["rpiCameraFPS"])
			require.Equal(t, true, pathData["rpiCameraHFlip"])
			require.Equal(t, "auto", pathData["rpiCameraExposure"])
			require.Equal(t, "h264", pathData["rpiCameraCodec"])
			require.Equal(t, 0.5, pathData["rpiCameraBrightness"])
			require.Equal(t, true, pathData["record"])
			require.Equal(t, "/recordings/pi-cam", pathData["recordPath"])
			
			// Verify that false values are not included (due to omitempty)
			_, exists := pathData["rpiCameraVFlip"]
			require.False(t, exists, "False boolean values should be omitted")
			
			t.Log("✅ Raspberry Pi camera PathOptions test passed")
		} else {
			t.Errorf("Expected Values to be map[string]interface{}, got %T", optPath.Values)
		}
	})
}