package cdp

import (
  "fmt"
  "testing"
  "time"
)

type HTask struct {
  name string
}

func (h *HTask) OnCdpEvent(msg *Message) {
  fmt.Println("======Event:", h.name, msg.Method)
}

func (h *HTask) OnCdpResp(msg *Message) {
  fmt.Println("======Resp:", h.name, msg.Method, msg.Id, msg.Result)
}

func TestTask(t *testing.T) {
  chrome, e := Launch("C:/Program Files (x86)/Google/Chrome/Application/chrome.exe")
  if e != nil {
    panic(e)
  }
  taskTB(chrome)
  taskJD(chrome)
  taskAmazon(chrome)
  time.Sleep(time.Second * 10)
}

func taskTB(chrome *Chrome) {
  h := &HTask{name: "TaoBao"}
  NewTask(chrome).
    Action(NewSimpleAction(Page.Enable, nil)).
    Action(NewSimpleAction(Page.Navigate, Param{"url": "https://item.taobao.com/item.htm?id=549226118434"})).
    Until(Page.LoadEventFired).
    Action(NewSimpleEvalAction("document.querySelector('#J_PromoPriceNum').textContent")).Run(h)
}

func taskJD(chrome *Chrome) {
  h := &HTask{name: "JingDong"}
  NewTask(chrome).
    Action(NewSimpleAction(Page.Enable, nil)).
    Action(NewSimpleAction(Page.Navigate, Param{"url": "https://item.jd.com/3693867.html"})).
    Until(Page.LoadEventFired).
    Action(NewSimpleEvalAction("document.querySelector('.J-p-3693867').textContent")).Run(h)
}

func taskAmazon(chrome *Chrome) {
  h := &HTask{name: "Amazon"}
  NewTask(chrome).
    Action(NewSimpleAction(Page.Enable, nil)).
    Action(NewSimpleAction(Page.Navigate, Param{"url": "https://www.amazon.cn/dp/B072RBZ7T1/"})).
    Until(Page.LoadEventFired).
    Action(NewSimpleEvalAction("document.querySelector('.a-color-price').textContent")).Run(h)
}
