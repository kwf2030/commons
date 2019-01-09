package pipeline

type HandlerContext struct {
  prev     *HandlerContext
  next     *HandlerContext
  pipeline *Pipeline
  name     string
  handler  Handler
}

func (ctx *HandlerContext) Prev() *HandlerContext {
  ctx.pipeline.mu.RLock()
  ret := ctx.prev
  if ret == ctx.pipeline.head {
    ret = nil
  }
  ctx.pipeline.mu.RUnlock()
  return ret
}

func (ctx *HandlerContext) Next() *HandlerContext {
  ctx.pipeline.mu.RLock()
  ret := ctx.next
  if ret == ctx.pipeline.tail {
    ret = nil
  }
  ctx.pipeline.mu.RUnlock()
  return ret
}

func (ctx *HandlerContext) Pipeline() *Pipeline {
  return ctx.pipeline
}

func (ctx *HandlerContext) Name() string {
  return ctx.name
}

func (ctx *HandlerContext) Handler() Handler {
  return ctx.handler
}

func (ctx *HandlerContext) Fire(data interface{}) {
  ctx.pipeline.mu.RLock()
  next := ctx.next
  ctx.pipeline.mu.RUnlock()
  if next != nil && next.handler != nil {
    next.handler.Handle(next, data)
  }
}

type Handler interface {
  Handle(*HandlerContext, interface{})
}

type defaultHandler struct{}

func (*defaultHandler) Handle(ctx *HandlerContext, data interface{}) {
  ctx.Fire(data)
}
