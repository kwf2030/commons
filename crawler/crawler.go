package crawler

import (
  "fmt"
  "html"
  "os"
  "runtime"
  "strings"
  "time"

  "github.com/kwf2030/commons/cdp"
  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/times"
  "github.com/rs/zerolog"
)

var (
  logFile  *os.File
  logger   *zerolog.Logger
  logLevel = zerolog.Disabled

  chrome *cdp.Chrome

  queue = make(chan []*Page, 1024)

  notify = make(chan []*Page, 2)
)

func SetLogLevel(level string) {
  switch strings.ToLower(level) {
  case "debug":
    logLevel = zerolog.DebugLevel
  case "info":
    logLevel = zerolog.InfoLevel
  case "warn":
    logLevel = zerolog.WarnLevel
  case "error":
    logLevel = zerolog.ErrorLevel
  }
  initLogger()
}

func initLogger() {
  now := times.Now()
  if logger == nil {
    next := now.Add(time.Hour * 24)
    next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
    time.AfterFunc(next.Sub(now), func() {
      go initLogger()
    })
  }
  zerolog.SetGlobalLevel(logLevel)
  zerolog.TimeFieldFormat = ""
  if logFile != nil {
    logFile.Close()
  }
  logFile, _ = os.Create(fmt.Sprintf("crawler_%s.log", now.Format(times.DateFormat4)))
  lg := zerolog.New(logFile).Level(logLevel).With().Timestamp().Logger()
  logger = &lg
}

func LaunchChrome(bin string, args ...string) error {
  if bin == "" {
    switch runtime.GOOS {
    case "windows":
      bin = "C:/Program Files (x86)/Google/Chrome/Application/chrome.exe"
    case "linux":
      bin = "/usr/bin/google-chrome-stable"
    }
  }
  c, e := cdp.Launch(bin, args...)
  if e != nil {
    return e
  }
  chrome = c
  return nil
}

func GetRules() *Rules {
  return rules
}

func Start() <-chan []*Page {
  if logger == nil {
    SetLogLevel("")
  }
  go func() {
    for pages := range queue {
      for _, p := range pages {
        if p == nil {
          continue
        }
        r := crawl(p)
        if len(r) > 0 {
          p.Result = r
        }
      }
      notify <- pages
    }
  }()
  return notify
}

func Stop() {
  if logFile != nil {
    logFile.Close()
  }
  if chrome != nil {
    tab, e := chrome.NewTab()
    if e == nil {
      tab.CallAsync(cdp.Browser.Close, nil)
    }
  }
}

func Enqueue(pages []*Page) {
  queue <- pages
}

func crawl(page *Page) map[string]interface{} {
  if page.Url == "" {
    return nil
  }
  if page.Group == "" {
    page.Group = "default"
  }
  addr := html.UnescapeString(page.Url)
  rule := rules.match(page.Group, addr)
  if rule == nil {
    return nil
  }
  ret := make(map[string]interface{}, len(rule.Fields))
  done := make(chan struct{})
  tab, _ := chrome.NewTab()
  tab.Subscribe(cdp.Page.LoadEventFired)
  tab.Call(cdp.Page.Enable, nil)
  tab.Call(cdp.Page.Navigate, cdp.Param{"url": addr})
  go func() {
    params := cdp.Param{"objectGroup": "console", "includeCommandLineAPI": true}
    for msg := range tab.C {
      if msg.Method != cdp.Page.LoadEventFired {
        continue
      }
      if rule.Prepare != nil && rule.Prepare.Eval != "" {
        params["expression"] = rule.Prepare.Eval
        r := tab.Call(cdp.Runtime.Evaluate, params)
        s := conv.String(conv.Map(r.Result, "result"), "value")
        if s == "" || s == "false" {
          close(done)
          return
        }
        if rule.Prepare.waitWhenReady > 0 {
          time.Sleep(rule.Prepare.waitWhenReady)
        }
      }
      for _, field := range rule.Fields {
        switch {
        case field.Eval != "" && field.Value != "":
          params["expression"] = fmt.Sprintf("{let value='%s';%s}", field.Value, field.Eval)
          if field.Export {
            r := tab.Call(cdp.Runtime.Evaluate, params)
            s := conv.String(conv.Map(r.Result, "result"), "value")
            ret[field.Name] = s

            // 定义JS变量
            params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, s)
            tab.CallAsync("cdp.Runtime.Evaluate", params)
          } else {
            tab.CallAsync(cdp.Runtime.Evaluate, params)
          }

        case field.Value != "":
          ret[field.Name] = field.Value

          // 定义JS变量
          params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, field.Value)
          tab.CallAsync("cdp.Runtime.Evaluate", params)

        case field.Eval != "":
          params["expression"] = field.Eval
          if field.Export {
            r := tab.Call(cdp.Runtime.Evaluate, params)
            s := conv.String(conv.Map(r.Result, "result"), "value")
            ret[field.Name] = s

            // 定义JS变量
            params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, s)
            tab.CallAsync("cdp.Runtime.Evaluate", params)
          } else {
            tab.CallAsync(cdp.Runtime.Evaluate, params)
          }
        }
        if field.wait > 0 {
          time.Sleep(field.wait)
        }
      }
      break
    }
    close(done)
  }()
  select {
  case <-time.After(rule.pageLoadTimeout):
    tab.C <- &cdp.Message{Method: cdp.Page.LoadEventFired}
    <-done
  case <-done:
  }
  tab.Close()
  /*if rule.Loop == nil || rule.Loop.Eval == "" {
    tab.Close()
  } else {
    go crawlLoop(rule, tab)
  }*/
  return ret
}

func crawlLoop(rule *rule, tab *cdp.Tab) {
  defer tab.Close()
  params := cdp.Param{"objectGroup": "console", "includeCommandLineAPI": true}
  if rule.Loop.Prepare != nil && rule.Loop.Prepare.Eval != "" {
    params["expression"] = rule.Loop.Prepare.Eval
    r := tab.Call(cdp.Runtime.Evaluate, params)
    s := conv.String(conv.Map(r.Result, "result"), "value")
    if s == "" || s == "false" {
      return
    }
    if rule.Loop.Prepare.waitWhenReady > 0 {
      time.Sleep(rule.Loop.Prepare.waitWhenReady)
    }
  }
  i := 0
  for {
    fmt.Println("loop:", i)

    // eval
    params["expression"] = rule.Loop.Eval
    r1 := tab.Call(cdp.Runtime.Evaluate, params)
    s1 := conv.String(conv.Map(r1.Result, "result"), "value")
    fmt.Println("eval")

    // 定义JS变量
    exp := fmt.Sprintf("index=%d;last=%s", i, s1)
    if i == 0 {
      exp = fmt.Sprintf("let index=%d;last=%s", i, s1)
    }
    params["expression"] = exp
    tab.CallAsync("cdp.Runtime.Evaluate", params)
    fmt.Println("var")

    // break
    if rule.Loop.Break != "" {
      params["expression"] = rule.Loop.Break
      r2 := tab.Call(cdp.Runtime.Evaluate, params)
      s2 := conv.String(conv.Map(r2.Result, "result"), "value")
      if s2 == "" || s2 == "false" {
        break
      }
    }
    fmt.Println("break")

    // next
    if rule.Loop.Next != "" {
      params["expression"] = rule.Loop.Next
      tab.CallAsync(cdp.Runtime.Evaluate, params)
      fmt.Println("next")
    }

    i++
    if i%rule.Loop.ExportCycle == 0 {
      // Export
      fmt.Println("export:", s1)
    }

    // wait
    if rule.Loop.wait > 0 {
      time.Sleep(rule.Loop.wait)
      fmt.Println("wait")
    }
  }
}

type Page struct {
  Id     string
  Url    string
  Group  string
  Result map[string]interface{}
}
