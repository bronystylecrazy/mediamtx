package externalcmd

import (
	"github.com/bluenviron/mediamtx/internal/externalcmd"
)

// Pool is a pool of external commands.
type Pool = externalcmd.Pool

// OnExitFunc is called when a command exits.
type OnExitFunc = externalcmd.OnExitFunc

// Environment is a map of environment variables.
type Environment = externalcmd.Environment

// Cmd is an external command.
type Cmd = externalcmd.Cmd

// NewCmd creates a new external command.
var NewCmd = externalcmd.NewCmd