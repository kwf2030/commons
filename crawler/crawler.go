package crawler

import (
  "fmt"
  "html"
  "runtime"
  "sync"
  "time"

  "github.com/kwf2030/commons/cdp"
  "github.com/kwf2030/commons/conv"
)

var chrome *cdp.Chrome

type Handler interface {
  OnFields(*Page, map[string]string)
  OnLoop(*Page, int, []string)
}

type Page struct {
  Id    string
  Url   string
  Group string

  rule    *rule
  tab     *cdp.Tab
  handler Handler

  lastCallId int32
}

func NewPage(id, url, group string) *Page {
  if url == "" {
    return nil
  }
  return &Page{Id: id, Url: url, Group: group}
}

func (page *Page) Crawl(h Handler) {
  if page.Url == "" {
    return
  }
  if page.Group == "" {
    page.Group = "default"
  }
  addr := html.UnescapeString(page.Url)
  rule := Rules.match(page.Group, addr)
  if rule == nil {
    return
  }
  tab, e := chrome.NewTab(page)
  if e != nil {
    return
  }
  page.rule = rule
  page.tab = tab
  page.handler = h
  tab.Subscribe(cdp.Page.LoadEventFired)
  tab.Call(cdp.Page.Enable, nil, nil)
  tab.Call(cdp.Page.Navigate, cdp.Param{"url": addr}, nil)
}

func (page *Page) OnCdpEvent(msg *cdp.Message) {
  if msg.Method == cdp.Page.LoadEventFired {
    page.crawlFields()
  }
}

func (page *Page) OnCdpResp(msg *cdp.Message) {

}

func (page *Page) crawlFields() {
  params := cdp.Param{"objectGroup": "console", "includeCommandLineAPI": true}
  rule := page.rule
  if rule.Prepare != nil {
    if rule.Prepare.Eval != "" {
      if rule.Prepare.Eval[0] == '{' {
        params["expression"] = rule.Prepare.Eval
      } else {
        params["expression"] = "{" + rule.Prepare.Eval + "}"
      }
      page.tab.Call(cdp.Runtime.Evaluate, params)
      /*s := conv.String(conv.Map(r.Result, "result"), "value")
      if s == "" || s == "false" {
        close(done)
        return
      }*/
    }
    if rule.Prepare.waitWhenReady > 0 {
      time.Sleep(rule.Prepare.waitWhenReady)
    }
  }
  /*if rule.Loop != nil && rule.Loop.Eval != "" && loopHandler != nil {
      crawlLoop(tab, rule, page, loopHandler)
    }*/
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
  var e error
  chrome, e = cdp.Launch(bin, args...)
  if e != nil {
    return e
  }
  return nil
}

func ExitChrome() {
  if chrome != nil {
    chrome.Exit()
  }
}

// 批量抓取会忽略Loop规则，确保Page.ID的唯一性
func CrawlBatch(pages []*Page) map[string]map[string]interface{} {
  if len(pages) == 0 {
    return nil
  }
  ret := make(map[string]map[string]interface{}, len(pages))
  handler := func(page *Page, result map[string]interface{}) {
    ret[page.Id] = result
  }
  for _, page := range pages {
    Crawl(page, handler, nil)
  }
  return ret
}

func crawlFields(tab *cdp.Tab, rule *rule) map[string]interface{} {
  ret := make(map[string]interface{}, len(rule.Fields))
  done := make(chan struct{})
  go func() {
    once := sync.Once{}
    for msg := range tab.C {
      if msg.Method != cdp.Page.LoadEventFired {
        continue
      }
      once.Do(func() {
        go func() {
          params := cdp.Param{"objectGroup": "console", "includeCommandLineAPI": true}
          if rule.Prepare != nil && rule.Prepare.Eval != "" {
            if rule.Prepare.Eval[0] == '{' {
              params["expression"] = rule.Prepare.Eval
            } else {
              params["expression"] = "{" + rule.Prepare.Eval + "}"
            }
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
              if !field.Export {
                tab.CallAsync(cdp.Runtime.Evaluate, params)
              } else {
                r := tab.Call(cdp.Runtime.Evaluate, params)
                s := conv.String(conv.Map(r.Result, "result"), "value")
                ret[field.Name] = s
                params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, s)
                tab.Call(cdp.Runtime.Evaluate, params)
              }

            case field.Value != "":
              ret[field.Name] = field.Value
              params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, field.Value)
              tab.Call(cdp.Runtime.Evaluate, params)

            case field.Eval != "":
              if field.Eval[0] == '{' {
                params["expression"] = field.Eval
              } else {
                params["expression"] = "{" + field.Eval + "}"
              }
              if !field.Export {
                tab.CallAsync(cdp.Runtime.Evaluate, params)
              } else {
                r := tab.Call(cdp.Runtime.Evaluate, params)
                s := conv.String(conv.Map(r.Result, "result"), "value")
                ret[field.Name] = s
                params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, s)
                // 不能用CallAsync，速度太快的话定义的变量还没生效
                tab.Call(cdp.Runtime.Evaluate, params)
              }
            }
            if field.wait > 0 {
              time.Sleep(field.wait)
            }
          }
          close(done)
        }()
      })
    }
  }()
  select {
  case <-time.After(rule.pageLoadTimeout):
    tab.C <- &cdp.Message{Method: cdp.Page.LoadEventFired}
    // 等待eval完成
    <-done
  case <-done:
  }
  return ret
}

func crawlLoop(tab *cdp.Tab, rule *rule, page *Page, handler func(*Page, int, []string)) {
  params := cdp.Param{"objectGroup": "console", "includeCommandLineAPI": true}
  if rule.Loop.Prepare != nil && rule.Loop.Prepare.Eval != "" {
    if rule.Loop.Prepare.Eval[0] == '{' {
      params["expression"] = rule.Loop.Prepare.Eval
    } else {
      params["expression"] = "{" + rule.Loop.Prepare.Eval + "}"
    }
    r := tab.Call(cdp.Runtime.Evaluate, params)
    s := conv.String(conv.Map(r.Result, "result"), "value")
    if s == "" || s == "false" {
      return
    }
    if rule.Loop.Prepare.waitWhenReady > 0 {
      time.Sleep(rule.Loop.Prepare.waitWhenReady)
    }
  }
  if rule.Loop.Eval[0] != '{' {
    rule.Loop.Eval = "{" + rule.Loop.Eval + "}"
  }
  if rule.Loop.Break != "" && rule.Loop.Break[0] != '{' {
    rule.Loop.Break = "{" + rule.Loop.Break + "}"
  }
  if rule.Loop.Next != "" && rule.Loop.Next[0] != '{' {
    rule.Loop.Next = "{" + rule.Loop.Next + "}"
  }
  i := 0
  bp := make([]string, rule.Loop.ExportCycle)
  for {
    // eval
    params["expression"] = rule.Loop.Eval
    r := tab.Call(cdp.Runtime.Evaluate, params)
    s := conv.String(conv.Map(r.Result, "result"), "value")

    exp := fmt.Sprintf("count=%d;last='%s'", i, s)
    if i == 0 {
      exp = fmt.Sprintf("let count=%d;last='%s'", i, s)
    }
    params["expression"] = exp
    tab.Call(cdp.Runtime.Evaluate, params)

    // break
    if rule.Loop.Break != "" {
      params["expression"] = rule.Loop.Break
      r := tab.Call(cdp.Runtime.Evaluate, params)
      s := conv.String(conv.Map(r.Result, "result"), "value")
      if s == "" || s == "false" {
        break
      }
    }

    // next
    if rule.Loop.Next != "" {
      params["expression"] = rule.Loop.Next
      tab.CallAsync(cdp.Runtime.Evaluate, params)
    }

    i++
    n := i % rule.Loop.ExportCycle
    if n == 0 {
      handler(page, i, bp)
    } else {
      bp[n-1] = s
    }

    // wait
    if rule.Loop.wait > 0 {
      time.Sleep(rule.Loop.wait)
    }
  }
}
