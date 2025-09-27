package hooks

import (
	"github.com/bluenviron/mediamtx/internal/hooks"
)

// OnReadyParams contains parameters for the OnReady hook.
type OnReadyParams = hooks.OnReadyParams

// OnConnectParams contains parameters for the OnConnect hook.
type OnConnectParams = hooks.OnConnectParams

// OnReadParams contains parameters for the OnRead hook.
type OnReadParams = hooks.OnReadParams

// OnDemandParams contains parameters for the OnDemand hook.
type OnDemandParams = hooks.OnDemandParams

// OnInitParams contains parameters for the OnInit hook.
type OnInitParams = hooks.OnInitParams

// OnReady is called when a stream becomes ready.
var OnReady = hooks.OnReady

// OnConnect is called when a client connects.
var OnConnect = hooks.OnConnect

// OnRead is called when a client starts reading.
var OnRead = hooks.OnRead

// OnDemand is called when a stream is requested on demand.
var OnDemand = hooks.OnDemand

// OnInit is called when the server initializes.
var OnInit = hooks.OnInit