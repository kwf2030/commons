package cdp

import (
  "time"
)

const defaultEvent = "DEFAULT"

type Action interface {
  Method() string
  Param() Param
}

type SimpleAction struct {
  method string
  param  Param
}

func NewSimpleAction(method string, param Param) *SimpleAction {
  if method == "" {
    return nil
  }
  return &SimpleAction{method: method, param: param}
}

func (sa *SimpleAction) Method() string {
  return sa.method
}

func (sa *SimpleAction) Param() Param {
  return sa.param
}

type waitAction time.Duration

func (wa waitAction) Method() string {
  return ""
}

func (wa waitAction) Param() Param {
  return nil
}

type EvalAction interface {
  Action
  Expressions() []string
}

type SimpleEvalAction struct {
  expressions []string
}

func NewSimpleEvalAction(expressions ...string) *SimpleEvalAction {
  if len(expressions) == 0 {
    return nil
  }
  return &SimpleEvalAction{expressions: expressions}
}

func (sea *SimpleEvalAction) Method() string {
  return Runtime.Evaluate
}

func (sea *SimpleEvalAction) Param() Param {
  return Param{"objectGroup": "console", "includeCommandLineAPI": true}
}

func (sea *SimpleEvalAction) Expressions() []string {
  return sea.expressions
}

type Task struct {
  chrome *Chrome
  tab    *Tab

  // 一个Domain事件对应多个Action（DomainEvent-->[]Action），
  // 没有事件的Action的key为DEFAULT
  actions map[string][]Action

  // 当前事件（用于链式调用）
  evt string

  handler Handler
}

func NewTask(c *Chrome) *Task {
  if c == nil {
    return nil
  }
  t := &Task{
    chrome:  c,
    actions: make(map[string][]Action, 2),
    evt:     defaultEvent,
  }
  t.actions[defaultEvent] = make([]Action, 0, 2)
  return t
}

func (t *Task) OnCdpEvent(msg *Message) {
  if actions, ok := t.actions[msg.Method]; ok {
    for _, action := range actions {
      t.runAction(action)
    }
  } else {
    if t.handler != nil {
      t.handler.OnCdpEvent(msg)
    }
  }
}

func (t *Task) OnCdpResp(msg *Message) bool {
  if t.handler != nil {
    return t.handler.OnCdpResp(msg)
  }
  return true
}

func (t *Task) CloseTab(msg *Message) {
  t.tab.Close()
}

func (t *Task) ExitChrome(msg *Message) {
  t.chrome.Exit()
}

func (t *Task) Action(action Action) *Task {
  if action != nil {
    t.actions[t.evt] = append(t.actions[t.evt], action)
  }
  return t
}

func (t *Task) Until(event string) *Task {
  if event != "" {
    if _, ok := t.actions[event]; !ok {
      t.evt = event
      t.actions[event] = make([]Action, 0, 16)
    }
  }
  return t
}

func (t *Task) Wait(duration time.Duration) *Task {
  if duration > 0 {
    t.Action(waitAction(duration))
  }
  return t
}

func (t *Task) Run(h Handler) *Task {
  tab, e := t.chrome.NewTab(t)
  if e != nil {
    return t
  }
  t.tab = tab
  t.handler = h
  for event := range t.actions {
    if event != defaultEvent {
      tab.Subscribe(event)
    }
  }
  for _, action := range t.actions[defaultEvent] {
    t.runAction(action)
  }
  return t
}

func (t *Task) runAction(action Action) {
  switch a := action.(type) {
  case waitAction:
    time.Sleep(time.Duration(a))
  case EvalAction:
    param := action.Param()
    if param == nil {
      param = Param{}
    }
    for _, expr := range a.Expressions() {
      if expr != "" {
        param["expression"] = expr
        t.tab.Call(action.Method(), param)
      }
    }
  default:
    t.tab.Call(action.Method(), action.Param())
  }
}
