package cdp

import (
  "context"
  "fmt"
  "sync"
  "testing"
  "time"
)

var wg sync.WaitGroup

func TestTab(t *testing.T) {
  chrome, e := Launch("C:/Program Files (x86)/Google/Chrome/Application/chrome.exe")
  if e != nil {
    panic(e)
  }
  wg.Add(3)
  tabTaoBao(chrome)
  tabJD(chrome)
  tabAmazon(chrome)
  wg.Wait()
}

func tabTaoBao(chrome *Chrome) {
  tab, e := chrome.NewTab()
  if e != nil {
    panic(e)
  }
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable, nil)
  tab.Call(Page.Navigate, Param{"url": "https://item.taobao.com/item.htm?id=549226118434"})
  go func() {
    for msg := range tab.C {
      if msg.Method == Page.LoadEventFired {
        v := tab.Call(Runtime.Evaluate, Param{"returnByValue": true, "expression": "document.querySelector('#J_PromoPriceNum').textContent"})
        fmt.Printf("%+v\n", v)
        tab.Close()
        wg.Done()
        break
      }
    }
  }()
}

func tabJD(chrome *Chrome) {
  tab, e := chrome.NewTab()
  if e != nil {
    panic(e)
  }
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable, nil)
  tab.Call(Page.Navigate, Param{"url": "https://item.jd.com/3693867.html"})
  go func() {
    for msg := range tab.C {
      if msg.Method == Page.LoadEventFired {
        v := tab.Call(Runtime.Evaluate, Param{"returnByValue": true, "expression": "document.querySelector('.J-p-3693867').textContent"})
        fmt.Printf("%+v\n", v)
        tab.Close()
        wg.Done()
        break
      }
    }
  }()
}

func tabAmazon(chrome *Chrome) {
  tab, e := chrome.NewTab()
  if e != nil {
    panic(e)
  }
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable, nil)
  tab.Call(Page.Navigate, Param{"url": "https://www.amazon.cn/dp/B072RBZ7T1/"})
  go func() {
    for msg := range tab.C {
      if msg.Method == Page.LoadEventFired {
        v := tab.Call(Runtime.Evaluate, Param{"returnByValue": true, "expression": "document.querySelector('.a-color-price').textContent"})
        fmt.Printf("%+v\n", v)
        tab.Close()
        wg.Done()
        break
      }
    }
  }()
}

func TestTask(t *testing.T) {
  chrome, e := Launch("C:/Program Files (x86)/Google/Chrome/Application/chrome.exe")
  if e != nil {
    panic(e)
  }
  wg.Add(3)
  taskTaoBao(chrome)
  taskJD(chrome)
  taskAmazon(chrome)
  wg.Wait()
}

func taskTaoBao(chrome *Chrome) {
  ctx, f := context.WithCancel(context.Background())
  NewTask(chrome).
    Action(NewAction(Page.Enable, nil)).
    Action(NewAction(Page.Navigate, Param{"url": "https://item.taobao.com/item.htm?id=549226118434"})).
    Until(Page.LoadEventFired).
    Action(NewEvalAction("document.querySelector('#J_PromoPriceNum').textContent", func(result Result) error {
      fmt.Println("HandleResult:\n", result)
      return nil
    })).Run(ctx)
  time.AfterFunc(time.Second*5, func() {
    f()
    wg.Done()
  })
}

func taskJD(chrome *Chrome) {
  ctx, f := context.WithCancel(context.Background())
  NewTask(chrome).
    Action(NewAction(Page.Enable, nil)).
    Action(NewAction(Page.Navigate, Param{"url": "https://item.jd.com/3693867.html"})).
    Until(Page.LoadEventFired).
    Action(NewEvalAction("document.querySelector('.J-p-3693867').textContent", func(result Result) error {
      fmt.Println("HandleResult:\n", result)
      return nil
    })).Run(ctx)
  time.AfterFunc(time.Second*5, func() {
    f()
    wg.Done()
  })
}

func taskAmazon(chrome *Chrome) {
  ctx, f := context.WithCancel(context.Background())
  NewTask(chrome).
    Action(NewAction(Page.Enable, nil)).
    Action(NewAction(Page.Navigate, Param{"url": "https://www.amazon.cn/dp/B072RBZ7T1/"})).
    Until(Page.LoadEventFired).
    Action(NewEvalAction("document.querySelector('.a-color-price').textContent", func(result Result) error {
      fmt.Println("HandleResult:\n", result)
      return nil
    })).Run(ctx)
  time.AfterFunc(time.Second*5, func() {
    f()
    wg.Done()
  })
}
