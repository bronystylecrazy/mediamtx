# MediaMTX API Usage Examples

## ✅ **Correct Usage for AddPathConfig**

The panic you encountered was due to incorrect usage of `OptionalPath`. Here's the correct way to use the MediaMTX API:

### **❌ Incorrect Usage (causes panic):**
```go
// This is WRONG - PathConf doesn't exist
server.AddPathConfig("testtt", &mediamtx.OptionalPath{
    Values: mediamtx.PathConf{  // ← PathConf type doesn't exist
        Name:   "testtt",
        Source: "rtsp://localhost:8554/test",
    },
})
```

### **✅ Correct Usage:**

#### **Option 1: Use the helper function (Recommended)**
```go
import "github.com/bluenviron/mediamtx/pkg/mediamtx"

// Create MediaMTX API
core := &mediamtx.Core{}
api := mediamtx.NewMediaMTXAPI(core)

// Method 1: Simple path with just a source
optPath := mediamtx.NewOptionalPath("rtsp://localhost:8554/test")
err := api.AddPathConfig("testtt", optPath)
if err != nil {
    log.Fatal(err)
}

// Method 2: Comprehensive path configuration using type-safe struct
options := mediamtx.PathOptions{
    Source:                     "rtsp://localhost:8554/test",
    Record:                     true,
    RecordPath:                 "/recordings",
    RecordFormat:               "mp4",
    MaxReaders:                 10,
    SourceOnDemand:             true,
    SourceOnDemandStartTimeout: "10s",
    RTSPTransport:              "udp",
    RTSPAnyPort:                true,
    UseAbsoluteTimestamp:       true,
    RunOnInit:                  "/scripts/init.sh",
    RunOnReady:                 "/scripts/ready.sh",
}
optPath := mediamtx.NewOptionalPathWithOptions(options)
err := api.AddPathConfig("testtt", optPath)
```

#### **Option 2: Manual OptionalPath creation**
```go
import (
    "github.com/bluenviron/mediamtx/pkg/mediamtx"
    "github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
)

// Create a Path instance manually
path := &conf.Path{
    Source: "rtsp://localhost:8554/test",
    Record: true,
}

// Wrap it in OptionalPath  
optPath := &conf.OptionalPath{
    Values: path,
}

// Add to configuration
err := api.AddPathConfig("testtt", optPath)
```

## **Complete Example**

```go
package main

import (
    "log"
    "github.com/bluenviron/mediamtx/pkg/mediamtx"
)

func main() {
    // Initialize MediaMTX
    core := &mediamtx.Core{}
    api := mediamtx.NewMediaMTXAPI(core)
    
    // Add a simple path
    optPath := mediamtx.NewOptionalPath("rtsp://localhost:8554/test")
    err := api.AddPathConfig("simple_path", optPath)
    if err != nil {
        log.Printf("Error adding simple path: %v", err)
    }
    
    // Add a comprehensive path configuration for camera recording
    recordingOptions := mediamtx.PathOptions{
        Source:                     "rtsp://camera1:8554/stream",
        Record:                     true,
        RecordPath:                 "/recordings/camera1",
        RecordFormat:               "mp4",
        RecordPartDuration:         "1h",
        RecordDeleteAfter:          "168h", // 7 days
        MaxReaders:                 5,
        SourceOnDemand:             false,
        RTSPTransport:              "tcp",
        RunOnRecordSegmentComplete: "/scripts/process_recording.sh",
        UseAbsoluteTimestamp:       true,
    }
    recordingPath := mediamtx.NewOptionalPathWithOptions(recordingOptions)
    err = api.AddPathConfig("camera1", recordingPath)
    if err != nil {
        log.Printf("Error adding recording path: %v", err)
    }
    
    // List all path configurations
    pagination := mediamtx.DefaultPagination()
    pathList, err := api.ListPathConfigs(pagination)
    if err != nil {
        log.Printf("Error listing paths: %v", err)
    } else {
        log.Printf("Found %d paths", pathList.ItemCount)
    }
}
```

## **Why the Helper Functions Fix the Panic**

The original panic occurred because:

1. **`PathConf` type doesn't exist** - You were trying to use a non-existent type
2. **`OptionalPath.Values` requires specific reflection setup** - The internal MediaMTX code expects `Values` to be properly structured using Go's reflection system
3. **Type mismatch in pagination** - When the wrong type was passed, the `paginateSlice` function received an unexpected struct instead of a slice pointer

The helper functions (`NewOptionalPath` and `NewOptionalPathWithOptions`) properly:
- Create valid `conf.Path` instances 
- Set up the reflection correctly
- Ensure type compatibility with the MediaMTX internals

## **Available Helper Functions**

### `NewOptionalPath(source string) *conf.OptionalPath`
Creates a simple path with just a source URL.

### `NewOptionalPathWithOptions(options PathOptions) *conf.OptionalPath`  
Creates a path with comprehensive configuration options using a type-safe struct based on the complete OpenAPI PathConf schema.

#### **Complete Field Coverage (90+ options available):**

**General Configuration:**
- `Source`, `SourceFingerprint`, `SourceOnDemand`, `SourceOnDemandStartTimeout`, `SourceOnDemandCloseAfter`
- `MaxReaders`, `SRTReadPassphrase`, `Fallback`, `UseAbsoluteTimestamp`

**Recording Configuration:**
- `Record`, `RecordPath`, `RecordFormat`, `RecordPartDuration`, `RecordMaxPartSize`
- `RecordSegmentDuration`, `RecordDeleteAfter`

**Publisher Source:**
- `OverridePublisher`, `SRTPublishPassphrase`

**RTSP Source Configuration:**
- `RTSPTransport`, `RTSPAnyPort`, `RTSPRangeType`, `RTSPRangeStart`, `RTSPUDPReadBufferSize`

**MPEG-TS & RTP Configuration:**
- `MPEGTSUDPReadBufferSize`, `RTPSDP`, `RTPUDPReadBufferSize`

**Redirect Source:**
- `SourceRedirect`

**Raspberry Pi Camera (40+ options):**
- `RPICameraCamID`, `RPICameraWidth`, `RPICameraHeight`, `RPICameraHFlip`, `RPICameraVFlip`
- `RPICameraBrightness`, `RPICameraContrast`, `RPICameraSaturation`, `RPICameraSharpness`
- `RPICameraExposure`, `RPICameraAWB`, `RPICameraCodec`, `RPICameraBitrate`, etc.

**Hooks & Scripts:**
- `RunOnInit`, `RunOnInitRestart`, `RunOnDemand`, `RunOnDemandRestart`
- `RunOnReady`, `RunOnReadyRestart`, `RunOnNotReady`, `RunOnRead`, `RunOnUnread`
- `RunOnRecordSegmentCreate`, `RunOnRecordSegmentComplete`

**Example with advanced options:**
```go
options := mediamtx.PathOptions{
    // Basic config
    Source: "rtsps://secure-camera:8554/stream",
    Record: true,
    RecordPath: "/recordings/secure",
    RecordFormat: "mp4",
    
    // Advanced RTSP
    RTSPTransport: "tcp",
    RTSPAnyPort: true,
    
    // Hooks
    RunOnReady: "/scripts/notify_ready.sh",
    RunOnRecordSegmentComplete: "/scripts/process_segment.sh",
    
    // Raspberry Pi Camera (if applicable)
    RPICameraWidth: 1920,
    RPICameraHeight: 1080,
    RPICameraFPS: 30.0,
    RPICameraHFlip: true,
}
```

## **Benefits of Type-Safe Struct Approach**

✅ **Compile-time safety** - Typos and wrong types caught at compile time  
✅ **IDE autocompletion** - Better development experience  
✅ **No reflection overhead** - Direct field assignment  
✅ **Clear documentation** - Struct fields are self-documenting  
✅ **Version compatibility** - Breaking changes visible at compile time