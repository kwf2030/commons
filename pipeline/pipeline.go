package pipeline

import (
  "fmt"
  "strings"
  "sync"
)

type Pipeline struct {
  head *HandlerContext
  tail *HandlerContext
  len  int
  mu   *sync.RWMutex
}

func New() *Pipeline {
  p := &Pipeline{
    head: &HandlerContext{name: "_head", handler: &defaultHandler{}},
    tail: &HandlerContext{name: "_tail", handler: &defaultHandler{}},
    mu:   &sync.RWMutex{},
  }
  p.head.pipeline = p
  p.head.next = p.tail
  p.tail.pipeline = p
  p.tail.prev = p.head
  return p
}

func (p *Pipeline) Run(data interface{}) {
  p.head.handler.Handle(p.head, data)
}

func (p *Pipeline) AddFirst(name string, h Handler) *Pipeline {
  if h != nil {
    p.mu.Lock()
    ctx := &HandlerContext{pipeline: p, name: name, handler: h}
    ctx.prev = p.head
    ctx.next = p.head.next
    ctx.next.prev = ctx
    p.head.next = ctx
    p.len++
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) AddLast(name string, h Handler) *Pipeline {
  if h != nil {
    p.mu.Lock()
    ctx := &HandlerContext{pipeline: p, name: name, handler: h}
    ctx.prev = p.tail.prev
    ctx.next = p.tail
    ctx.prev.next = ctx
    p.tail.prev = ctx
    p.len++
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) AddBefore(mark, name string, h Handler) *Pipeline {
  if mark != "" && name != "" && h != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, name: name, handler: h}
      ctx.prev = markCtx.prev
      ctx.next = markCtx
      ctx.prev.next = ctx
      markCtx.prev = ctx
      p.len++
    }
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) AddAfter(mark, name string, h Handler) *Pipeline {
  if mark != "" && name != "" && h != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, name: name, handler: h}
      ctx.prev = markCtx
      ctx.next = markCtx.next
      ctx.next.prev = ctx
      markCtx.next = ctx
      p.len++
    }
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) Clear() {
  p.mu.Lock()
  p.head.next = p.tail
  p.tail.prev = p.head
  p.len = 0
  p.mu.Unlock()
}

func (p *Pipeline) Remove(name string) *Pipeline {
  if name != "" {
    p.mu.Lock()
    if ctx := p.GetUnsafe(name); ctx != nil {
      ctx.prev.next = ctx.next
      ctx.next.prev = ctx.prev
      p.len--
    }
    p.mu.Unlock()
  }
  return p
}

func (p *Pipeline) Replace(mark, name string, h Handler) *Pipeline {
  if mark != "" && name != "" && h != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &HandlerContext{pipeline: p, name: name, handler: h}
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

func (p *Pipeline) Get(name string) *HandlerContext {
  if name == "" {
    return nil
  }
  p.mu.RLock()
  ctx := p.GetUnsafe(name)
  p.mu.RUnlock()
  return ctx
}

func (p *Pipeline) GetUnsafe(name string) *HandlerContext {
  if p.len == 0 {
    return nil
  }
  for ctx := p.head.next; ctx != nil && ctx != p.tail; ctx = ctx.next {
    if ctx.name == name {
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

func (p *Pipeline) String() string {
  p.mu.RLock()
  sb := strings.Builder{}
  i := 1
  for ctx := p.head.next; ctx != nil && ctx != p.tail; ctx = ctx.next {
    fmt.Fprintf(&sb, "[%d]%s(%T)\n", i, ctx.name, ctx.handler)
    i++
  }
  p.mu.RUnlock()
  return sb.String()
}
