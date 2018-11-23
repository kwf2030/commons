package crawler

import (
  "errors"
  "fmt"
  "html"
  "runtime"
  "sync"
  "time"

  "github.com/kwf2030/commons/cdp"
  "github.com/kwf2030/commons/conv"
)

var (
  ErrInvalidArgs   = errors.New("invalid args")
  ErrNoMatchedRule = errors.New("no matched rule")

  chrome *cdp.Chrome
)

type Handler interface {
  OnFields(*Page, map[string]string)

  OnLoop(*Page, int, []string)

  OnComplete(*Page)
}

type Page struct {
  Id    string
  Url   string
  Group string

  rule    *rule
  tab     *cdp.Tab
  handler Handler

  once sync.Once
}

func NewPage(id, url, group string) *Page {
  if url == "" {
    return nil
  }
  return &Page{Id: id, Url: url, Group: group}
}

func (p *Page) OnCdpEvent(msg *cdp.Message) {
  if msg.Method == cdp.Page.LoadEventFired {
    // 如果超时，就有可能存在两次回调（超时一次回调和正常一次回调），
    // once就是为了防止重复调用
    p.once.Do(func() {
      m := p.crawlFields()
      if p.handler != nil {
        p.handler.OnFields(p, m)
      }
      if p.rule.Loop != nil {
        p.crawlLoop()
      }
      if p.handler != nil {
        p.handler.OnComplete(p)
      }
    })
  }
}

func (p *Page) OnCdpResp(msg *cdp.Message) bool {
  return false
}

func (p *Page) Crawl(h Handler) error {
  if p.Url == "" {
    return ErrInvalidArgs
  }
  if p.Group == "" {
    p.Group = "default"
  }
  addr := html.UnescapeString(p.Url)
  rule := Rules.match(p.Group, addr)
  if rule == nil {
    return ErrNoMatchedRule
  }
  tab, e := chrome.NewTab(p)
  if e != nil {
    return e
  }
  p.rule = rule
  p.tab = tab
  p.handler = h
  tab.Subscribe(cdp.Page.LoadEventFired)
  tab.Call(cdp.Page.Enable, nil)
  tab.Call(cdp.Page.Navigate, cdp.Param{"url": addr})
  // todo 大量定时器，如果有性能问题改用时间轮
  time.AfterFunc(p.rule.pageLoadTimeout, func() {
    tab.FireEvent(cdp.Page.LoadEventFired, nil)
  })
  return nil
}

func (p *Page) crawlFields() map[string]string {
  rule := p.rule
  ret := make(map[string]string, len(rule.Fields))
  params := cdp.Param{"objectGroup": "console", "includeCommandLineAPI": true}
  if rule.Prepare != nil {
    if rule.Prepare.Eval != "" {
      if rule.Prepare.Eval[0] == '{' {
        params["expression"] = rule.Prepare.Eval
      } else {
        params["expression"] = "{" + rule.Prepare.Eval + "}"
      }
      _, ch := p.tab.Call(cdp.Runtime.Evaluate, params)
      msg := <-ch
      r := conv.String(conv.Map(msg.Result, "result"), "value")
      if r != "true" {
        return ret
      }
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
        p.tab.Call(cdp.Runtime.Evaluate, params)
      } else {
        _, ch := p.tab.Call(cdp.Runtime.Evaluate, params)
        msg := <-ch
        r := conv.String(conv.Map(msg.Result, "result"), "value")
        ret[field.Name] = r
        params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, r)
        p.tab.Call(cdp.Runtime.Evaluate, params)
      }

    case field.Value != "":
      ret[field.Name] = field.Value
      params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, field.Value)
      p.tab.Call(cdp.Runtime.Evaluate, params)

    case field.Eval != "":
      if field.Eval[0] == '{' {
        params["expression"] = field.Eval
      } else {
        params["expression"] = "{" + field.Eval + "}"
      }
      if !field.Export {
        p.tab.Call(cdp.Runtime.Evaluate, params)
      } else {
        _, ch := p.tab.Call(cdp.Runtime.Evaluate, params)
        msg := <-ch
        r := conv.String(conv.Map(msg.Result, "result"), "value")
        ret[field.Name] = r
        params["expression"] = fmt.Sprintf("const %s='%s'", field.Name, r)
        p.tab.Call(cdp.Runtime.Evaluate, params)
      }
    }
    if field.wait > 0 {
      time.Sleep(field.wait)
    }
  }
  return ret
}

func (p *Page) crawlLoop() {
  rule := p.rule
  params := cdp.Param{"objectGroup": "console", "includeCommandLineAPI": true}
  if rule.Loop.Prepare != nil {
    if rule.Loop.Prepare.Eval != "" {
      if rule.Loop.Prepare.Eval[0] == '{' {
        params["expression"] = rule.Loop.Prepare.Eval
      } else {
        params["expression"] = "{" + rule.Loop.Prepare.Eval + "}"
      }
      _, ch := p.tab.Call(cdp.Runtime.Evaluate, params)
      msg := <-ch
      r := conv.String(conv.Map(msg.Result, "result"), "value")
      if r != "true" {
        return
      }
    }
    if rule.Loop.Prepare.waitWhenReady > 0 {
      time.Sleep(rule.Loop.Prepare.waitWhenReady)
    }
  }
  if rule.Loop.Eval != "" && rule.Loop.Eval[0] != '{' {
    rule.Loop.Eval = "{" + rule.Loop.Eval + "}"
  }
  if rule.Loop.Break != "" && rule.Loop.Break[0] != '{' {
    rule.Loop.Break = "{" + rule.Loop.Break + "}"
  }
  if rule.Loop.Next != "" && rule.Loop.Next[0] != '{' {
    rule.Loop.Next = "{" + rule.Loop.Next + "}"
  }
  var v string
  i := 1
  bp := make([]string, rule.Loop.ExportCycle)
  for {
    // eval
    if rule.Loop.Eval != "" {
      params["expression"] = rule.Loop.Eval
      _, ch := p.tab.Call(cdp.Runtime.Evaluate, params)
      msg := <-ch
      v = conv.String(conv.Map(msg.Result, "result"), "value")
      exp := fmt.Sprintf("count=%d;last='%s'", i, v)
      if i == 1 {
        exp = "let " + exp
      }
      params["expression"] = exp
      p.tab.Call(cdp.Runtime.Evaluate, params)
      n := i % rule.Loop.ExportCycle
      if n != 0 {
        bp[n-1] = v
      } else {
        bp[rule.Loop.ExportCycle-1] = v
        if p.handler != nil {
          p.handler.OnLoop(p, i, bp)
        }
      }
    }

    // break
    if rule.Loop.Break != "" {
      params["expression"] = rule.Loop.Break
      _, ch := p.tab.Call(cdp.Runtime.Evaluate, params)
      msg := <-ch
      r := conv.String(conv.Map(msg.Result, "result"), "value")
      if r == "true" {
        break
      }
    }

    // next
    if rule.Loop.Next != "" {
      params["expression"] = rule.Loop.Next
      p.tab.Call(cdp.Runtime.Evaluate, params)
    }

    // wait
    if rule.Loop.wait > 0 {
      time.Sleep(rule.Loop.wait)
    }

    i++
  }
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
/*func CrawlBatch(pages []*Page) map[string]map[string]interface{} {
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
}*/
