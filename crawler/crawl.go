package crawler

import (
  "fmt"
  "time"

  "github.com/kwf2030/commons/cdp"
  "github.com/kwf2030/commons/conv"
)

func crawl(addr string) map[string]interface{} {
  rule := allRules.match("default", addr)
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
        if field.Wait > 0 {
          time.Sleep(field.Wait)
        }
      }
      break
    }
    close(done)
  }()
  select {
  case <-time.After(rule.PageLoadTimeout):
    logger.Debug().Msg("crawl timeout, execute expression")
    tab.C <- &cdp.Message{Method: cdp.Page.LoadEventFired}
    <-done
  case <-done:
    logger.Debug().Msg("crawl done")
  }
  tab.Close()
  return ret
}
