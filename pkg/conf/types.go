package conf

import (
	"github.com/bluenviron/mediamtx/internal/conf"
)

// RecordFormat is a recording format.
type RecordFormat = conf.RecordFormat

// Recording formats.
const (
	RecordFormatFMP4   = conf.RecordFormatFMP4
	RecordFormatMPEGTS = conf.RecordFormatMPEGTS
)

// RTSPRangeType is a RTSP range type.
type RTSPRangeType = conf.RTSPRangeType

// RTSP range types.
const (
	RTSPRangeTypeUndefined = conf.RTSPRangeTypeUndefined
	RTSPRangeTypeClock     = conf.RTSPRangeTypeClock
	RTSPRangeTypeNPT       = conf.RTSPRangeTypeNPT
	RTSPRangeTypeSMPTE     = conf.RTSPRangeTypeSMPTE
)

// RTSPTransport is a RTSP transport configuration.
type RTSPTransport = conf.RTSPTransport

// RTSPTransports is a map of RTSP transport protocols.
type RTSPTransports = conf.RTSPTransports

// RTSPAuthMethods is a list of RTSP authentication methods.
type RTSPAuthMethods = conf.RTSPAuthMethods

// Duration is a time.Duration wrapper with JSON/YAML support.
type Duration = conf.Duration

// StringSize is a uint64 wrapper for size configurations.
type StringSize = conf.StringSize

// LogLevel is a logger level wrapper.
type LogLevel = conf.LogLevel

// LogDestinations is a list of log destinations.
type LogDestinations = conf.LogDestinations

// IPNetworks is a list of IP networks.
type IPNetworks = conf.IPNetworks

// WebRTCICEServer is a WebRTC ICE server configuration.
type WebRTCICEServer = conf.WebRTCICEServer

// WebRTCICEServers is a list of WebRTC ICE servers.
type WebRTCICEServers = conf.WebRTCICEServers

// HLSVariant is a HLS variant configuration.
type HLSVariant = conf.HLSVariant

// Encryption is an encryption type.
type Encryption = conf.Encryption

// Encryption types.
const (
	EncryptionNo       = conf.EncryptionNo
	EncryptionOptional = conf.EncryptionOptional
	EncryptionStrict   = conf.EncryptionStrict
)