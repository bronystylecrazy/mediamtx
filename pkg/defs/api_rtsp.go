package defs

import (
	"github.com/bluenviron/mediamtx/internal/defs"
)

// APIRTSPConn is a RTSP connection.
type APIRTSPConn = defs.APIRTSPConn

// APIRTSPConnsList is a list of RTSP connections.
type APIRTSPConnsList = defs.APIRTSPConnsList

// APIRTSPSessionState is the state of a RTSP session.
type APIRTSPSessionState = defs.APIRTSPSessionState

// RTSP session states.
const (
	APIRTSPSessionStateIdle    = defs.APIRTSPSessionStateIdle
	APIRTSPSessionStateRead    = defs.APIRTSPSessionStateRead
	APIRTSPSessionStatePublish = defs.APIRTSPSessionStatePublish
)

// APIRTSPSession is a RTSP session.
type APIRTSPSession = defs.APIRTSPSession

// APIRTSPSessionList is a list of RTSP sessions.
type APIRTSPSessionList = defs.APIRTSPSessionList