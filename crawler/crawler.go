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
    e := os.MkdirAll("log", os.ModePerm)
    if e != nil {
      return
    }
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
        if p.Url == "" {
          continue
        }
        r := crawl(html.UnescapeString(p.Url))
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

func crawl(addr string) map[string]interface{} {
  rule := rules.match("default", addr)
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
      }
      for _, field := range rule.Fields {
        switch {
        case field.Eval != "" && field.Value != "":
          params["expression"] = fmt.Sprintf("{let value=%s;%s}", field.Value, field.Eval)
          if field.Export {
            r := tab.Call(cdp.Runtime.Evaluate, params)
            s := conv.String(conv.Map(r.Result, "result"), "value")
            ret[field.Name] = s

            // 定义JS变量
            params["expression"] = fmt.Sprintf("const %s=%s", field.Name, s)
            tab.CallAsync("cdp.Runtime.Evaluate", params)
          } else {
            tab.CallAsync(cdp.Runtime.Evaluate, params)
          }

        case field.Value != "":
          ret[field.Name] = field.Value

          // 定义JS变量
          params["expression"] = fmt.Sprintf("const %s=%s", field.Name, field.Value)
          tab.CallAsync("cdp.Runtime.Evaluate", params)

        case field.Eval != "":
          params["expression"] = field.Eval
          if field.Export {
            r := tab.Call(cdp.Runtime.Evaluate, params)
            s := conv.String(conv.Map(r.Result, "result"), "value")
            ret[field.Name] = s

            // 定义JS变量
            params["expression"] = fmt.Sprintf("const %s=%s", field.Name, s)
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
  case <-time.After(rule.pageLoadTimeout * 100):
    logger.Debug().Msg("crawl timeout, execute expression")
    tab.C <- &cdp.Message{Method: cdp.Page.LoadEventFired}
    <-done
  case <-done:
    logger.Debug().Msg("crawl done")
  }
  tab.Close()
  return ret
}

type Page struct {
  Id     string
  Url    string
  Result map[string]interface{}
}
