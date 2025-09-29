package defs

import (
	"github.com/bluenviron/mediamtx/internal/defs"
)

// APISRTConnState is the state of a SRT connection.
type APISRTConnState = defs.APISRTConnState

// SRT connection states.
const (
	APISRTConnStateIdle    = defs.APISRTConnStateIdle
	APISRTConnStateRead    = defs.APISRTConnStateRead
	APISRTConnStatePublish = defs.APISRTConnStatePublish
)

// APISRTConn is a SRT connection.
type APISRTConn = defs.APISRTConn

// APISRTConnList is a list of SRT connections.
type APISRTConnList = defs.APISRTConnList