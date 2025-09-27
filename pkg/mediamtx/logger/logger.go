package logger

import (
	"github.com/bluenviron/mediamtx/internal/logger"
)

// Level is a log level.
type Level = logger.Level

// Log levels.
const (
	Debug = logger.Debug
	Info  = logger.Info
	Warn  = logger.Warn
	Error = logger.Error
)

// Writer is an entity that can write logs.
type Writer = logger.Writer

// Logger is the main logger implementation.
type Logger = logger.Logger

// Destination is a log destination.
type Destination = logger.Destination

// Log destinations.
const (
	DestinationStdout = logger.DestinationStdout
	DestinationFile   = logger.DestinationFile
	DestinationSyslog = logger.DestinationSyslog
)

// New creates a new logger.
var New = logger.New