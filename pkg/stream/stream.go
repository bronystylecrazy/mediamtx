package stream

import (
	"github.com/bluenviron/mediamtx/internal/stream"
)

// Reader is an entity that can read from a stream.
type Reader = stream.Reader

// ReadFunc is the callback passed to AddReader().
type ReadFunc = stream.ReadFunc

// Stream is a media stream.
type Stream = stream.Stream