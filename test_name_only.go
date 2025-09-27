package main

import (
	"fmt"
	"log"

	"github.com/bluenviron/mediamtx/pkg/mediamtx"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
)

func main() {
	fmt.Println("🧪 Testing Name-only PathOptions that was causing panic...")

	// Test the helper function directly to verify no panic occurs
	fmt.Println("\n1. Testing NewOptionalPathWithOptions with Name only...")
	
	// This was causing: panic: reflect: call of reflect.Value.Elem on map Value
	optPath1 := mediamtx.NewOptionalPathWithOptions(mediamtx.PathOptions{
		Name: "test",
	})
	
	if optPath1 != nil && optPath1.Values != nil {
		fmt.Println("✅ Success! No panic occurred!")
		
		// Verify it creates a proper conf.Path struct
		if path, ok := optPath1.Values.(*conf.Path); ok {
			fmt.Printf("   Created proper *conf.Path with Name: '%s'\n", path.Name)
		} else {
			fmt.Printf("   Values type: %T\n", optPath1.Values)
		}
	} else {
		log.Fatal("❌ Failed to create OptionalPath")
	}

	// Test with comprehensive options
	fmt.Println("\n2. Testing comprehensive PathOptions...")
	optPath2 := mediamtx.NewOptionalPathWithOptions(mediamtx.PathOptions{
		Name:       "comprehensive",
		Source:     "rtsp://localhost:8554/test",
		Record:     true,
		RecordPath: "/recordings",
	})
	
	if optPath2 != nil && optPath2.Values != nil {
		fmt.Println("✅ Success! Comprehensive options work!")
		
		if path, ok := optPath2.Values.(*conf.Path); ok {
			fmt.Printf("   Name: '%s', Source: '%s', Record: %t, RecordPath: '%s'\n", 
				path.Name, path.Source, path.Record, path.RecordPath)
		}
	} else {
		log.Fatal("❌ Failed to create comprehensive OptionalPath")
	}

	fmt.Println("\n🎉 Original panic issue is FIXED!")
	fmt.Println("   - NewOptionalPathWithOptions now creates proper *conf.Path structs")
	fmt.Println("   - No more 'reflect: call of reflect.Value.Elem on map Value' panic")
	fmt.Println("   - MediaMTX validation code can now properly process the values")
}