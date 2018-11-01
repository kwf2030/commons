package cdp

import (
  "context"
  "time"
)

const defaultEvent = "DEFAULT"

type Action interface {
  Method() string

  Param() Param
}

type EvalAction interface {
  Action

  Expressions() []string

  Handle(Result) error
}

type SimpleAction struct {
  method string
  param  Param
}

func NewAction(method string, param Param) *SimpleAction {
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

type SimpleEvalAction struct {
  expression string
  handler    func(Result) error
}

func NewEvalAction(expression string, handler func(Result) error) *SimpleEvalAction {
  return &SimpleEvalAction{expression: expression, handler: handler}
}

func (sea *SimpleEvalAction) Method() string {
  return Runtime.Evaluate
}

func (sea *SimpleEvalAction) Param() Param {
  return Param{"objectGroup": "console", "includeCommandLineAPI": true}
}

func (sea *SimpleEvalAction) Expressions() []string {
  return []string{sea.expression}
}

func (sea *SimpleEvalAction) Handle(result Result) error {
  if sea.handler != nil {
    return sea.handler(result)
  }
  return nil
}

type Task struct {
  chrome Chrome

  // 一个Domain事件对应多个Action（DomainEvent-->[]Action），
  // 没有事件的Action的key为DEFAULT，会优先执行
  actions map[string][]Action

  // 当前事件
  evt string
}

func NewTask(c Chrome) *Task {
  t := &Task{
    chrome:  c,
    actions: make(map[string][]Action, 4),
    evt:     defaultEvent,
  }
  t.actions[defaultEvent] = make([]Action, 0, 2)
  return t
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
  return t.Action(waitAction(duration))
}

func (t *Task) Run(ctx context.Context) {
  tab, e := t.chrome.NewTab()
  if e != nil {
    return
  }
  for event := range t.actions {
    if event != defaultEvent {
      tab.Subscribe(event)
    }
  }
  for _, action := range t.actions[defaultEvent] {
    switch a := action.(type) {
    case waitAction:
      time.Sleep(time.Duration(a))
    case EvalAction:
      t.runEvalAction(tab, a)
    default:
      tab.CallAsync(action.Method(), action.Param())
    }
  }
  // 因为默认加了一个DEFAULT事件，所以Task.actions长度必定不小于1，
  // 如果大于1，说明有事件要监听（执行对应事件的Actions）
  if len(t.actions) > 1 {
    go t.runActions(ctx, tab)
  }
}

func (t *Task) runActions(ctx context.Context, tab *Tab) {
  for {
    select {
    case <-ctx.Done():
      tab.Close()
      return
    case msg := <-tab.C:
      if actions, ok := t.actions[msg.Method]; ok {
        for _, action := range actions {
          switch a := action.(type) {
          case waitAction:
            time.Sleep(time.Duration(a))
          case EvalAction:
            t.runEvalAction(tab, a)
          default:
            tab.CallAsync(action.Method(), action.Param())
          }
        }
      }
    }
  }
}

func (t *Task) runEvalAction(tab *Tab, action EvalAction) {
  param := action.Param()
  if param == nil {
    param = Param{}
  }
  for _, expr := range action.Expressions() {
    param["expression"] = expr
    e := action.Handle(tab.Call(action.Method(), param).Result)
    if e != nil {
      break
    }
  }
}
