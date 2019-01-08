package pipeline

type Op struct {
  What int
  Msg  string
  Data []byte
  Val  interface{}
  Err  error
}

type Handler interface {
  Handle(*HandlerContext, int, interface{})
}

type HandlerContext struct {
  prev     *HandlerContext
  next     *HandlerContext
  pipeline *Pipeline
  handler  Handler
}

func (ctx *HandlerContext) Pipeline() *Pipeline {
  return ctx.pipeline
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

func (ctx *HandlerContext) FireNext(what int, val interface{}) {
  ctx.pipeline.mu.RLock()
  next := ctx.next
  ctx.pipeline.mu.RUnlock()
  if next != nil && next.handler != nil {
    next.handler.Handle(next, what, val)
  }
}

type defaultHandler struct{}

func (*defaultHandler) Handle(ctx *HandlerContext, what int, val interface{}) {
  ctx.FireNext(what, val)
}
