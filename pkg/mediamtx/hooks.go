package mediamtx

import "context"

type PathHook interface {
	OnDemand(context.Context, *PathHandler, string)
	OnUnDemand(context.Context, *PathHandler, string)
	OnDemandStatic(context.Context, *PathHandler, string)
	OnUnDemandStatic(context.Context, *PathHandler, string)
	OnInit(context.Context, *PathHandler)
	OnUnInit(context.Context, *PathHandler)
	OnReady(context.Context, *PathHandler)
	OnNotReady(context.Context, *PathHandler)
}
