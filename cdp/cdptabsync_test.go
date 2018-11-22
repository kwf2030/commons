package cdp

import (
  "fmt"
  "sync"
  "testing"
)

var wgSync sync.WaitGroup

type HSync struct {
  name   string
  expr   string
  ch     chan struct{}
  tab    *Tab
  callId int32
}

func (h *HSync) OnCdpEvent(msg *Message) {
  fmt.Println("======Event:", h.name, msg.Method)
  if msg.Method == Page.LoadEventFired {
    id := h.tab.Call(Runtime.Evaluate, Param{"returnByValue": true, "expression": h.expr})
    fmt.Println("call id:", id)
    h.callId = id
  }
}

func (h *HSync) OnCdpResp(msg *Message) {
  fmt.Println("======Resp:", h.name, msg.Method, msg.Id, msg.Result)
  if msg.Id == h.callId {
    h.ch <- struct{}{}
    h.tab.Close()
    wgSync.Done()
  }
}

func TestTabSync(t *testing.T) {
  chrome, e := Launch("C:/Program Files (x86)/Google/Chrome/Application/chrome.exe")
  if e != nil {
    panic(e)
  }
  wgSync.Add(3)
  // 如果cap为零最后一次会阻塞
  ch := make(chan struct{}, 1)
  go func() {
    fs := []func(*Chrome, chan struct{}){tabSyncTB, tabSyncJD, tabSyncAmazon}
    for _, f := range fs {
      <-ch
      f(chrome, ch)
    }
  }()
  ch <- struct{}{}
  wgSync.Wait()
  chrome.Exit()
}

func tabSyncTB(chrome *Chrome, ch chan struct{}) {
  h := &HSync{name: "TaoBao", expr: "document.querySelector('#J_PromoPriceNum').textContent", ch: ch}
  tab, e := chrome.NewTab(h)
  if e != nil {
    panic(e)
  }
  h.tab = tab
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable, nil)
  tab.Call(Page.Navigate, Param{"url": "https://item.taobao.com/item.htm?id=549226118434"})
}

func tabSyncJD(chrome *Chrome, ch chan struct{}) {
  h := &HSync{name: "JingDong", expr: "document.querySelector('.J-p-3693867').textContent", ch: ch}
  tab, e := chrome.NewTab(h)
  if e != nil {
    panic(e)
  }
  h.tab = tab
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable, nil)
  tab.Call(Page.Navigate, Param{"url": "https://item.jd.com/3693867.html"})
}

func tabSyncAmazon(chrome *Chrome, ch chan struct{}) {
  h := &HSync{name: "Amazon", expr: "document.querySelector('.a-color-price').textContent", ch: ch}
  tab, e := chrome.NewTab(h)
  if e != nil {
    panic(e)
  }
  h.tab = tab
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable, nil)
  tab.Call(Page.Navigate, Param{"url": "https://www.amazon.cn/dp/B072RBZ7T1/"})
}
