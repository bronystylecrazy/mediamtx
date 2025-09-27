package mediamtx

import (
	"context"
	"testing"
	"time"
)

func TestAddPathConfigDoesNotBlock(t *testing.T) {
	ctx := context.Background()
	
	// Create a Core instance without starting it
	core, err := New(Options{})
	if err != nil {
		t.Fatalf("Failed to create core: %v", err)
	}
	defer core.Close(ctx)

	// Create API instance  
	api := NewMediaMTXAPI(core)

	// Create a test path configuration
	pathConf := NewOptionalPathWithOptions(PathOptions{
		Name:   "test-blocking",
		Source: "publisher",
	})

	// This should complete quickly even when Core.Run() hasn't been called
	done := make(chan error, 1)
	go func() {
		done <- api.AddPathConfig("test-blocking", pathConf)
	}()

	// Wait for completion with timeout
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("AddPathConfig failed: %v", err)
		}
		t.Logf("✓ AddPathConfig completed successfully without blocking")
	case <-time.After(2 * time.Second):
		t.Fatal("❌ AddPathConfig blocked for more than 2 seconds - this indicates the fix didn't work")
	}

	// Verify the path was added to the configuration
	paths := core.Conf.Paths
	if paths["test-blocking"] == nil {
		t.Fatal("Path was not added to configuration")
	}

	t.Logf("✓ Path successfully added to configuration with name: test-blocking")
}

func TestAddPathConfigWithRunningCore(t *testing.T) {
	ctx := context.Background()
	
	// Create a Core instance and start it
	core, err := New(Options{})
	if err != nil {
		t.Fatalf("Failed to create core: %v", err)
	}
	defer core.Close(ctx)

	// Start the Core in a goroutine so it can process events
	go func() {
		if err := core.Run(ctx); err != nil {
			t.Logf("Core.Run() error: %v", err)
		}
	}()

	// Give Core time to start
	time.Sleep(100 * time.Millisecond)

	// Create API instance  
	api := NewMediaMTXAPI(core)

	// Create a test path configuration
	pathConf := NewOptionalPathWithOptions(PathOptions{
		Name:   "test-running",
		Source: "publisher",
	})

	// This should also complete quickly when Core is running
	done := make(chan error, 1)
	go func() {
		done <- api.AddPathConfig("test-running", pathConf)
	}()

	// Wait for completion with timeout
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("AddPathConfig failed: %v", err)
		}
		t.Logf("✓ AddPathConfig completed successfully with running Core")
	case <-time.After(2 * time.Second):
		t.Fatal("❌ AddPathConfig blocked for more than 2 seconds")
	}

	// Verify the path was added to the configuration
	paths := core.Conf.Paths
	if paths["test-running"] == nil {
		t.Fatal("Path was not added to configuration")
	}

	t.Logf("✓ Path successfully added to running Core with name: test-running")
}