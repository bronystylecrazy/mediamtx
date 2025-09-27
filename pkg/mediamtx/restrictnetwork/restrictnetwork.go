package restrictnetwork

import (
	"github.com/bluenviron/mediamtx/internal/restrictnetwork"
)

// Restrict prevents listening on IPv6 when address is 0.0.0.0.
var Restrict = restrictnetwork.Restrict