package wechatbot

import "sync"

type Pipeline struct {
  head, tail *HandlerContext
  len        int
  m          *sync.RWMutex
}

func newPipeline() *Pipeline {
  p := &Pipeline{head: &HandlerContext{}, tail: &HandlerContext{}, m: &sync.RWMutex{}}
  p.head.pipeline = p
  p.head.next = p.tail
  p.tail.pipeline = p
  p.tail.prev = p.head
  return p
}

func (p *Pipeline) AddFirst(h Handler) *Pipeline {
  if h != nil {
    p.m.Lock()
    ctx := &HandlerContext{pipeline: p, handler: h}
    ctx.prev = p.head
    ctx.next = p.head.next
    p.head.next.prev = ctx
    p.head.next = ctx
    p.len++
    p.m.Unlock()
  }
  return p
}

func (p *Pipeline) AddLast(h Handler) *Pipeline {
  if h != nil {
    p.m.Lock()
    ctx := &HandlerContext{pipeline: p, handler: h}
    ctx.prev = p.tail.prev
    ctx.next = p.tail
    p.tail.prev.next = ctx
    p.tail.prev = ctx
    p.len++
    p.m.Unlock()
  }
  return p
}

func (p *Pipeline) AddBefore(h, mark Handler) *Pipeline {
  if h != nil && mark != nil {
    p.m.Lock()
    if markCtx := p.get(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, handler: h}
      ctx.prev = markCtx.prev
      ctx.next = markCtx
      markCtx.prev = ctx
      p.len++
    }
    p.m.Unlock()
  }
  return p
}

func (p *Pipeline) AddAfter(h, mark Handler) *Pipeline {
  if h != nil && mark != nil {
    p.m.Lock()
    if markCtx := p.get(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, handler: h}
      ctx.prev = markCtx
      ctx.next = markCtx.next
      markCtx.next = ctx
      p.len++
    }
    p.m.Unlock()
  }
  return p
}

func (p *Pipeline) Remove(h Handler) *Pipeline {
  if h != nil {
    p.m.Lock()
    if ctx := p.get(h); ctx != nil {
      ctx.prev.next = ctx.next
      ctx.next.prev = ctx.prev
      p.len--
    }
    p.m.Unlock()
  }
  return p
}

func (p *Pipeline) Replace(h, mark Handler) *Pipeline {
  if h != nil && mark != nil {
    p.m.Lock()
    if markCtx := p.get(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, handler: h}
      ctx.prev = markCtx.prev
      ctx.next = markCtx.next
      ctx.prev.next = ctx
      ctx.next.prev = ctx
    }
    p.m.Unlock()
  }
  return p
}

func (p *Pipeline) First() *HandlerContext {
  p.m.RLock()
  defer p.m.RUnlock()
  if p.len == 0 {
    return nil
  }
  return p.head.next
}

func (p *Pipeline) Last() *HandlerContext {
  p.m.RLock()
  defer p.m.RUnlock()
  if p.len == 0 {
    return nil
  }
  return p.tail.prev
}

func (p *Pipeline) Get(h Handler) *HandlerContext {
  if h == nil {
    return nil
  }
  p.m.RLock()
  ctx := p.get(h)
  p.m.RUnlock()
  return ctx
}

func (p *Pipeline) get(h Handler) *HandlerContext {
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
  p.m.RLock()
  defer p.m.RUnlock()
  return p.len
}

func (p *Pipeline) Clear() {
  p.m.Lock()
  p.head.next = p.tail
  p.tail.prev = p.head
  p.len = 0
  p.m.Unlock()
}
