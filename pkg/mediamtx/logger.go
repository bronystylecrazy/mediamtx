package mediamtx

import "github.com/bluenviron/mediamtx/internal/logger"

type LogLevel = logger.Level

const (
	Debug LogLevel = iota + 1
	Info
	Warn
	Error
)
