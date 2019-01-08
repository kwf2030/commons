package pipeline

import "sync"

type Pipeline struct {
  head *HandlerContext
  tail *HandlerContext
  len  int
  mu   *sync.RWMutex
}

func New() *Pipeline {
  p := &Pipeline{
    head: &HandlerContext{handler: &defaultHandler{}},
    tail: &HandlerContext{handler: &defaultHandler{}},
    mu:   &sync.RWMutex{},
  }
  p.head.pipeline = p
  p.head.next = p.tail
  p.tail.pipeline = p
  p.tail.prev = p.head
  return p
}

func (p *Pipeline) Run(what int, val interface{}) {
  p.head.handler.Handle(p.head, what, val)
}

func (p *Pipeline) AddFirst(h Handler) *Pipeline {
  if h != nil {
    p.mu.Lock()
    ctx := &HandlerContext{pipeline: p, handler: h}
    ctx.prev = p.head
    ctx.next = p.head.next
    p.head.next.prev = ctx
    p.head.next = ctx
    p.len++
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) AddLast(h Handler) *Pipeline {
  if h != nil {
    p.mu.Lock()
    ctx := &HandlerContext{pipeline: p, handler: h}
    ctx.prev = p.tail.prev
    ctx.next = p.tail
    p.tail.prev.next = ctx
    p.tail.prev = ctx
    p.len++
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) AddBefore(h, mark Handler) *Pipeline {
  if h != nil && mark != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, handler: h}
      ctx.prev = markCtx.prev
      ctx.next = markCtx
      markCtx.prev = ctx
      p.len++
    }
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) AddAfter(h, mark Handler) *Pipeline {
  if h != nil && mark != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, handler: h}
      ctx.prev = markCtx
      ctx.next = markCtx.next
      markCtx.next = ctx
      p.len++
    }
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) Remove(h Handler) *Pipeline {
  if h != nil {
    p.mu.Lock()
    if ctx := p.GetUnsafe(h); ctx != nil {
      ctx.prev.next = ctx.next
      ctx.next.prev = ctx.prev
      p.len--
    }
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) Replace(h, mark Handler) *Pipeline {
  if h != nil && mark != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, handler: h}
      ctx.prev = markCtx.prev
      ctx.next = markCtx.next
      ctx.prev.next = ctx
      ctx.next.prev = ctx
    }
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) First() *HandlerContext {
  p.mu.RLock()
  var ctx *HandlerContext
  if p.len != 0 {
    ctx = p.head.next
  }
  p.mu.RUnlock()
  return ctx
}

func (p *Pipeline) Last() *HandlerContext {
  p.mu.RLock()
  var ctx *HandlerContext
  if p.len != 0 {
    ctx = p.tail.prev
  }
  p.mu.RUnlock()
  return ctx
}

func (p *Pipeline) Get(h Handler) *HandlerContext {
  if h == nil {
    return nil
  }
  p.mu.RLock()
  ctx := p.GetUnsafe(h)
  p.mu.RUnlock()
  return ctx
}

func (p *Pipeline) GetUnsafe(h Handler) *HandlerContext {
  if p.len == 0 {
    return nil
  }
  for ctx := p.head.next; ctx != nil && ctx != p.tail; ctx = ctx.next {
    if ctx.handler == h {
      return ctx
    }
  }
  return nil
}

func (p *Pipeline) Len() int {
  p.mu.RLock()
  n := p.len
  p.mu.RUnlock()
  return n
}

func (p *Pipeline) Clear() {
  p.mu.Lock()
  p.head.next = p.tail
  p.tail.prev = p.head
  p.len = 0
  p.mu.Unlock()
}
