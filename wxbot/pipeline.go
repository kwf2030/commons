package wxbot

import (
  "fmt"
  "strings"
  "sync"
)

type pipeline struct {
  head *handlerCtx
  tail *handlerCtx
  len  int
  mu   *sync.RWMutex
}

func newPipeline() *pipeline {
  p := &pipeline{
    head: &handlerCtx{name: "_head", handler: &defaultHandler{}},
    tail: &handlerCtx{name: "_tail", handler: &defaultHandler{}},
    mu:   &sync.RWMutex{},
  }
  p.head.pipeline = p
  p.head.next = p.tail
  p.tail.pipeline = p
  p.tail.prev = p.head
  return p
}

func (p *pipeline) Fire(evt event) {
  p.head.handler.Handle(p.head, evt)
}

func (p *pipeline) AddFirst(name string, h handler) *pipeline {
  if h != nil {
    p.mu.Lock()
    ctx := &handlerCtx{pipeline: p, name: name, handler: h}
    ctx.prev = p.head
    ctx.next = p.head.next
    ctx.next.prev = ctx
    p.head.next = ctx
    p.len++
    p.mu.Unlock()
  }
  return p
}

func (p *pipeline) AddLast(name string, h handler) *pipeline {
  if h != nil {
    p.mu.Lock()
    ctx := &handlerCtx{pipeline: p, name: name, handler: h}
    ctx.prev = p.tail.prev
    ctx.next = p.tail
    ctx.prev.next = ctx
    p.tail.prev = ctx
    p.len++
    p.mu.Unlock()
  }
  return p
}

func (p *pipeline) AddBefore(mark, name string, h handler) *pipeline {
  if mark != "" && name != "" && h != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &handlerCtx{pipeline: p, name: name, handler: h}
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

func (p *pipeline) AddAfter(mark, name string, h handler) *pipeline {
  if mark != "" && name != "" && h != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &handlerCtx{pipeline: p, name: name, handler: h}
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

func (p *pipeline) Clear() {
  p.mu.Lock()
  p.head.next = p.tail
  p.tail.prev = p.head
  p.len = 0
  p.mu.Unlock()
}

func (p *pipeline) Remove(name string) *pipeline {
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

func (p *pipeline) Replace(mark, name string, h handler) *pipeline {
  if mark != "" && name != "" && h != nil {
    p.mu.Lock()
    if markCtx := p.GetUnsafe(mark); markCtx != nil {
      ctx := &handlerCtx{pipeline: p, name: name, handler: h}
      ctx.prev = markCtx.prev
      ctx.next = markCtx.next
      ctx.prev.next = ctx
      ctx.next.prev = ctx
    }
    p.mu.Unlock()
  }
  return p
}

func (p *pipeline) First() *handlerCtx {
  p.mu.RLock()
  var ctx *handlerCtx
  if p.len != 0 {
    ctx = p.head.next
  }
  p.mu.RUnlock()
  return ctx
}

func (p *pipeline) Last() *handlerCtx {
  p.mu.RLock()
  var ctx *handlerCtx
  if p.len != 0 {
    ctx = p.tail.prev
  }
  p.mu.RUnlock()
  return ctx
}

func (p *pipeline) Get(name string) *handlerCtx {
  if name == "" {
    return nil
  }
  p.mu.RLock()
  ctx := p.GetUnsafe(name)
  p.mu.RUnlock()
  return ctx
}

func (p *pipeline) GetUnsafe(name string) *handlerCtx {
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

func (p *pipeline) Len() int {
  p.mu.RLock()
  n := p.len
  p.mu.RUnlock()
  return n
}

func (p *pipeline) String() string {
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

type handlerCtx struct {
  prev     *handlerCtx
  next     *handlerCtx
  pipeline *pipeline
  name     string
  handler  handler
}

func (ctx *handlerCtx) Prev() *handlerCtx {
  ctx.pipeline.mu.RLock()
  ret := ctx.prev
  if ret == ctx.pipeline.head {
    ret = nil
  }
  ctx.pipeline.mu.RUnlock()
  return ret
}

func (ctx *handlerCtx) Next() *handlerCtx {
  ctx.pipeline.mu.RLock()
  ret := ctx.next
  if ret == ctx.pipeline.tail {
    ret = nil
  }
  ctx.pipeline.mu.RUnlock()
  return ret
}

func (ctx *handlerCtx) Pipeline() *pipeline {
  return ctx.pipeline
}

func (ctx *handlerCtx) Name() string {
  return ctx.name
}

func (ctx *handlerCtx) Handler() handler {
  return ctx.handler
}

func (ctx *handlerCtx) Fire(evt event) {
  ctx.pipeline.mu.RLock()
  next := ctx.next
  ctx.pipeline.mu.RUnlock()
  if next != nil {
    next.handler.Handle(next, evt)
  }
}

type handler interface {
  Handle(*handlerCtx, event)
}

type defaultHandler struct{}

func (*defaultHandler) Handle(ctx *handlerCtx, evt event) {
  ctx.Fire(evt)
}

type event struct {
  what int
  data []byte
  val  interface{}
  err  error
}
