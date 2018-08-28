package cdp

import (
  "context"
  "fmt"
  "sync"
  "testing"
  "time"
)

var wg sync.WaitGroup

// 在headless模式下，设置--user-data-dir会导致Chrome无响应（非headless没影响，可能是Chrome的bug），
// Chrome: 68.0.3440.106
func launch() Chrome {
  bin := "C:/Program Files (x86)/Google/Chrome/Application/chrome.exe"
  args := []string{
    "--remote-debugging-port=9222",
    "--headless",
    "--disable-gpu",
    "--no-first-run",
    "--no-default-browser-check",
    "--window-size=1024,768",
    "--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
  }
  chrome, _ := LaunchChrome(bin, args...)
  return chrome
}

func TestTab(t *testing.T) {
  wg.Add(3)
  chrome := launch()
  tabTaoBao(chrome)
  tabJD(chrome)
  tabAmazon(chrome)
  wg.Wait()
}

func tabTaoBao(chrome Chrome) {
  tab, _ := chrome.NewTab()
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable)
  tab.Call(Page.Navigate, Params{"url": "https://item.taobao.com/item.htm?id=549226118434"})
  go func() {
    for msg := range tab.C {
      if msg.Method == Page.LoadEventFired {
        v := tab.Call(Runtime.Evaluate, Params{"returnByValue": true, "expression": "document.querySelector('#J_PromoPriceNum').textContent"})
        fmt.Printf("%+v\n", v)
        tab.Close()
        wg.Done()
        break
      }
    }
  }()
}

func tabJD(chrome Chrome) {
  tab, _ := chrome.NewTab()
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable)
  tab.Call(Page.Navigate, Params{"url": "https://item.jd.com/3693867.html"})
  go func() {
    for msg := range tab.C {
      if msg.Method == Page.LoadEventFired {
        v := tab.Call(Runtime.Evaluate, Params{"returnByValue": true, "expression": "document.querySelector('.J-p-3693867').textContent"})
        fmt.Printf("%+v\n", v)
        tab.Close()
        wg.Done()
        break
      }
    }
  }()
}

func tabAmazon(chrome Chrome) {
  tab, _ := chrome.NewTab()
  tab.Subscribe(Page.LoadEventFired)
  tab.Call(Page.Enable)
  tab.Call(Page.Navigate, Params{"url": "https://www.amazon.cn/dp/B072RBZ7T1/"})
  go func() {
    for msg := range tab.C {
      if msg.Method == Page.LoadEventFired {
        v := tab.Call(Runtime.Evaluate, Params{"returnByValue": true, "expression": "document.querySelector('.a-color-price').textContent"})
        fmt.Printf("%+v\n", v)
        tab.Close()
        wg.Done()
        break
      }
    }
  }()
}

func TestTask(t *testing.T) {
  wg.Add(3)
  chrome := launch()
  taskTaoBao(chrome)
  taskJD(chrome)
  taskAmazon(chrome)
  wg.Wait()
}

func taskTaoBao(chrome Chrome) {
  ctx, f := context.WithCancel(context.Background())
  NewTask(chrome).
    Action(NewAction(Page.Enable)).
    Action(NewAction(Page.Navigate, Params{"url": "https://item.taobao.com/item.htm?id=549226118434"})).
    WaitFor(Page.LoadEventFired).
    Action(NewEvaluateAction("document.querySelector('#J_PromoPriceNum').textContent", func(result Result) error {
      fmt.Println("HandleResult:\n", result)
      return nil
    })).
    Run(ctx)
  time.AfterFunc(time.Second*8, func() {
    f()
    wg.Done()
  })
}

func taskJD(chrome Chrome) {
  ctx, f := context.WithCancel(context.Background())
  NewTask(chrome).
    Action(NewAction(Page.Enable)).
    Action(NewAction(Page.Navigate, Params{"url": "https://item.jd.com/3693867.html"})).
    WaitFor(Page.LoadEventFired).
    Action(NewEvaluateAction("document.querySelector('.J-p-3693867').textContent", func(result Result) error {
      fmt.Println("HandleResult:\n", result)
      return nil
    })).
    Run(ctx)
  time.AfterFunc(time.Second*8, func() {
    f()
    wg.Done()
  })
}

func taskAmazon(chrome Chrome) {
  ctx, f := context.WithCancel(context.Background())
  NewTask(chrome).
    Action(NewAction(Page.Enable)).
    Action(NewAction(Page.Navigate, Params{"url": "https://www.amazon.cn/dp/B072RBZ7T1/"})).
    WaitFor(Page.LoadEventFired).
    Action(NewEvaluateAction("document.querySelector('.a-color-price').textContent", func(result Result) error {
      fmt.Println("HandleResult:\n", result)
      return nil
    })).
    Run(ctx)
  time.AfterFunc(time.Second*8, func() {
    f()
    wg.Done()
  })
}
