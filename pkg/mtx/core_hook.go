package mtx

import "context"

func (p *Core) OnDemand(ctx context.Context, p2 *PathHandler, s string) {
	if p.opts.PathHook == nil {
		return
	}
	p.opts.PathHook.OnDemand(ctx, p2, s)
}

func (p *Core) OnUnDemand(ctx context.Context, p2 *PathHandler, s string) {
	if p.opts.PathHook == nil {
		return
	}
	p.opts.PathHook.OnUnDemand(ctx, p2, s)
}

func (p *Core) OnDemandStatic(ctx context.Context, p2 *PathHandler, s string) {
	if p.opts.PathHook == nil {
		return
	}
	p.opts.PathHook.OnDemandStatic(ctx, p2, s)
}

func (p *Core) OnUnDemandStatic(ctx context.Context, p2 *PathHandler, s string) {
	if p.opts.PathHook == nil {
		return
	}
	p.opts.PathHook.OnUnDemandStatic(ctx, p2, s)
}

func (p *Core) OnInit(ctx context.Context, p2 *PathHandler) {
	if p.opts.PathHook == nil {
		return
	}
	p.opts.PathHook.OnInit(ctx, p2)
}

func (p *Core) OnUnInit(ctx context.Context, p2 *PathHandler) {
	if p.opts.PathHook == nil {
		return
	}
	p.opts.PathHook.OnUnInit(ctx, p2)
}

func (p *Core) OnReady(ctx context.Context, p2 *PathHandler) {
	if p.opts.PathHook == nil {
		return
	}
	p.opts.PathHook.OnReady(ctx, p2)
}

func (p *Core) OnNotReady(ctx context.Context, p2 *PathHandler) {
	if p.opts.PathHook == nil {
		return
	}
	p.opts.PathHook.OnNotReady(ctx, p2)
}
