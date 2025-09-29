package mtx

import (
	"github.com/bluenviron/mediamtx/pkg/logger"
)

type LogLevel = logger.Level

const (
	Debug = logger.Debug
	Info  = logger.Info
	Warn  = logger.Warn
	Error = logger.Error
)
