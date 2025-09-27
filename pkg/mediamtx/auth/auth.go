package auth

import (
	"github.com/bluenviron/mediamtx/internal/auth"
)

// Credentials contains authentication credentials.
type Credentials = auth.Credentials

// Error is an authentication error.
type Error = auth.Error

// Manager is an authentication manager.
type Manager = auth.Manager

// Protocol is the protocol used to connect to the server.
type Protocol = auth.Protocol

// Request is an authentication request.
type Request = auth.Request

// Protocol constants.
const (
	ProtocolRTSP   = auth.ProtocolRTSP
	ProtocolRTMP   = auth.ProtocolRTMP
	ProtocolHLS    = auth.ProtocolHLS
	ProtocolWebRTC = auth.ProtocolWebRTC
	ProtocolSRT    = auth.ProtocolSRT
)

// PauseAfterError is the pause after authentication failure.
const PauseAfterError = auth.PauseAfterError