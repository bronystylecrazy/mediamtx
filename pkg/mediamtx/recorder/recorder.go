package recorder

import (
	"github.com/bluenviron/mediamtx/internal/recorder"
)

// OnSegmentCreateFunc is called when a segment is created.
type OnSegmentCreateFunc = recorder.OnSegmentCreateFunc

// OnSegmentCompleteFunc is called when a segment is complete.
type OnSegmentCompleteFunc = recorder.OnSegmentCompleteFunc

// Recorder is a recorder.
type Recorder = recorder.Recorder