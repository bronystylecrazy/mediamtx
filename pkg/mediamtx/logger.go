package mediamtx

import "github.com/bluenviron/mediamtx/pkg/mediamtx/logger"

type LogLevel = logger.Level

const (
	Debug = logger.Debug
	Info  = logger.Info
	Warn  = logger.Warn
	Error = logger.Error
)
