package defs

import (
	"github.com/bluenviron/mediamtx/internal/defs"
)

// Source is an entity that can provide a stream.
// it can be:
// - Publisher
// - staticsources.Handler
// - core.sourceRedirect
type Source = defs.Source

// FormatsToCodecs returns the name of codecs of given formats.
var FormatsToCodecs = defs.FormatsToCodecs

// FormatsInfo returns a description of formats.
var FormatsInfo = defs.FormatsInfo

// MediasToCodecs returns the name of codecs of given formats.
var MediasToCodecs = defs.MediasToCodecs

// MediasInfo returns a description of medias.
var MediasInfo = defs.MediasInfo