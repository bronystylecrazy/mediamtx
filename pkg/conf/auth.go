package conf

import (
	"github.com/bluenviron/mediamtx/internal/conf"
)

// AuthMethod is an authentication method.
type AuthMethod = conf.AuthMethod

// Authentication methods.
const (
	AuthMethodInternal = conf.AuthMethodInternal
	AuthMethodHTTP     = conf.AuthMethodHTTP
	AuthMethodJWT      = conf.AuthMethodJWT
)

// AuthAction is an authentication action.
type AuthAction = conf.AuthAction

// Authentication actions.
const (
	AuthActionPublish  AuthAction = "publish"
	AuthActionRead     AuthAction = "read"
	AuthActionPlayback AuthAction = "playback"
	AuthActionAPI      AuthAction = "api"
	AuthActionMetrics  AuthAction = "metrics"
	AuthActionPprof    AuthAction = "pprof"
)

// AuthInternalUser is an internal user.
type AuthInternalUser = conf.AuthInternalUser

// AuthInternalUsers is a list of internal users.
type AuthInternalUsers = conf.AuthInternalUsers

// AuthInternalUserPermission is a user permission.
type AuthInternalUserPermission = conf.AuthInternalUserPermission

// AuthInternalUserPermissions is a list of user permissions.
type AuthInternalUserPermissions = conf.AuthInternalUserPermissions

// Credential is a credential.
type Credential = conf.Credential
