package conf

import (
	"github.com/bluenviron/mediamtx/internal/conf"
)

// Conf is the main configuration struct.
type Conf = conf.Conf

// Path is the path configuration.
type Path = conf.Path

// Global is the global configuration.
type Global = conf.Global

// OptionalGlobal is an optional global configuration wrapper.
type OptionalGlobal = conf.OptionalGlobal

// OptionalPath is an optional path configuration wrapper.
type OptionalPath = conf.OptionalPath

// Load loads configuration from a file.
var Load = conf.Load

// IsValidPathName validates a path name.
var IsValidPathName = conf.IsValidPathName

// FindPathConf finds path configuration by name.
var FindPathConf = conf.FindPathConf

// ErrPathNotFound is returned when a path is not found.
var ErrPathNotFound = conf.ErrPathNotFound