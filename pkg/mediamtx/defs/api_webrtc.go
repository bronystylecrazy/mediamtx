package defs

import (
	"github.com/bluenviron/mediamtx/internal/defs"
)

// APIWebRTCSessionState is the state of a WebRTC connection.
type APIWebRTCSessionState = defs.APIWebRTCSessionState

// WebRTC session states.
const (
	APIWebRTCSessionStateRead    = defs.APIWebRTCSessionStateRead
	APIWebRTCSessionStatePublish = defs.APIWebRTCSessionStatePublish
)

// APIWebRTCSession is a WebRTC session.
type APIWebRTCSession = defs.APIWebRTCSession

// APIWebRTCSessionList is a list of WebRTC sessions.
type APIWebRTCSessionList = defs.APIWebRTCSessionList