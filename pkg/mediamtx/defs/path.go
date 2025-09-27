package defs

import (
	"github.com/bluenviron/mediamtx/internal/defs"
)

// PathNoStreamAvailableError is returned when no one is publishing.
type PathNoStreamAvailableError = defs.PathNoStreamAvailableError

// Path is a path.
type Path = defs.Path

// PathFindPathConfRes contains the response of FindPathConf().
type PathFindPathConfRes = defs.PathFindPathConfRes

// PathFindPathConfReq contains arguments of FindPathConf().
type PathFindPathConfReq = defs.PathFindPathConfReq

// PathDescribeRes contains the response of Describe().
type PathDescribeRes = defs.PathDescribeRes

// PathDescribeReq contains arguments of Describe().
type PathDescribeReq = defs.PathDescribeReq

// PathAddPublisherRes contains the response of AddPublisher().
type PathAddPublisherRes = defs.PathAddPublisherRes

// PathAddPublisherReq contains arguments of AddPublisher().
type PathAddPublisherReq = defs.PathAddPublisherReq

// PathRemovePublisherReq contains arguments of RemovePublisher().
type PathRemovePublisherReq = defs.PathRemovePublisherReq

// PathAddReaderRes contains the response of AddReader().
type PathAddReaderRes = defs.PathAddReaderRes

// PathAddReaderReq contains arguments of AddReader().
type PathAddReaderReq = defs.PathAddReaderReq

// PathRemoveReaderReq contains arguments of RemoveReader().
type PathRemoveReaderReq = defs.PathRemoveReaderReq

// PathSourceStaticSetReadyRes contains the response of SetReady().
type PathSourceStaticSetReadyRes = defs.PathSourceStaticSetReadyRes

// PathSourceStaticSetReadyReq contains arguments of SetReady().
type PathSourceStaticSetReadyReq = defs.PathSourceStaticSetReadyReq

// PathSourceStaticSetNotReadyReq contains arguments of SetNotReady().
type PathSourceStaticSetNotReadyReq = defs.PathSourceStaticSetNotReadyReq