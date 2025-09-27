package defs

import (
	"github.com/bluenviron/mediamtx/internal/defs"
)

// APIRTMPConnState is the state of a RTMP connection.
type APIRTMPConnState = defs.APIRTMPConnState

// RTMP connection states.
const (
	APIRTMPConnStateIdle    = defs.APIRTMPConnStateIdle
	APIRTMPConnStateRead    = defs.APIRTMPConnStateRead
	APIRTMPConnStatePublish = defs.APIRTMPConnStatePublish
)

// APIRTMPConn is a RTMP connection.
type APIRTMPConn = defs.APIRTMPConn

// APIRTMPConnList is a list of RTMP connections.
type APIRTMPConnList = defs.APIRTMPConnList