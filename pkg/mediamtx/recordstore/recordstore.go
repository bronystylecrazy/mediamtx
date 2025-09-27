package recordstore

import (
	"github.com/bluenviron/mediamtx/internal/recordstore"
)

// Segment is a recording segment.
type Segment = recordstore.Segment

// Path is a recording path.
type Path = recordstore.Path

// FindAllPathsWithSegments finds all paths containing recording segments.
var FindAllPathsWithSegments = recordstore.FindAllPathsWithSegments

// FindSegments finds recording segments for a path.
var FindSegments = recordstore.FindSegments

// PathAddExtension adds file extension based on recording format.
var PathAddExtension = recordstore.PathAddExtension

// CommonPath gets the common path prefix.
var CommonPath = recordstore.CommonPath

// ErrNoSegmentsFound is returned when no segments are found.
var ErrNoSegmentsFound = recordstore.ErrNoSegmentsFound