package mediamtx

import (
	"fmt"
	"time"
)

// ExampleDirectAPIUsage demonstrates how to use the DirectAPI without gin/HTTP dependencies
func ExampleDirectAPIUsage() {
	// Create a MediaMTX core instance
	core := &Core{}
	
	// Initialize the direct API
	api := NewDirectAPI(core)
	
	fmt.Println("=== MediaMTX Direct API Example ===")
	
	// 1. Configuration Management
	fmt.Println("\n1. Configuration Management:")
	defaults := api.GetPathDefaults()
	fmt.Printf("   Path defaults created: %v\n", defaults != nil)
	
	// 2. Pagination
	fmt.Println("\n2. Pagination Support:")
	pagination, err := PaginateFromStrings("10", "1")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Items per page: %d, Page: %d\n", pagination.ItemsPerPage, pagination.Page)
	}
	
	// 3. Authentication (with nil AuthManager)
	fmt.Println("\n3. Authentication:")
	authReq, err := api.CreateAuthRequest("testuser", "testpass", "", "127.0.0.1")
	if err != nil {
		fmt.Printf("   Error creating auth request: %v\n", err)
	} else {
		authErr := api.Authenticate(authReq)
		if authErr == nil {
			fmt.Println("   Authentication passed (nil AuthManager)")
		} else {
			fmt.Printf("   Authentication failed: %v\n", authErr)
		}
	}
	
	// 4. Recording Info Structure
	fmt.Println("\n4. Recording Information:")
	recordingInfo := &RecordingInfo{
		PathName:        "example_path",
		TotalRecordings: 3,
		TotalDuration:   5 * time.Minute,
	}
	fmt.Printf("   Path: %s, Recordings: %d, Duration: %v\n", 
		recordingInfo.PathName, recordingInfo.TotalRecordings, recordingInfo.TotalDuration)
	
	// 5. Server Management (would work with running servers)
	fmt.Println("\n5. Server Management:")
	fmt.Println("   Direct API provides methods for:")
	fmt.Println("   - RTSP/RTSPS connections and sessions")
	fmt.Println("   - RTMP/RTMPS connections") 
	fmt.Println("   - WebRTC sessions")
	fmt.Println("   - SRT connections")
	fmt.Println("   - HLS muxers")
	
	// 6. Recording Management
	fmt.Println("\n6. Recording Management:")
	fmt.Println("   Direct API provides methods for:")
	fmt.Println("   - Querying recordings with time filters")
	fmt.Println("   - Getting recording information")
	fmt.Println("   - Deleting recording segments")
	fmt.Println("   - Managing recording paths")
	
	fmt.Println("\n=== Direct API ready for use! ===")
}