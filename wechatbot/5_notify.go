package wechatbot

import (
  "bytes"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "net/url"

  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const statusNotifyUrlPath = "/webwxstatusnotify"

const opNotify = 0x5001

// 在手机上显示"已登录Web微信"
type statusNotifyReq struct {
  req *req
}

func (r *statusNotifyReq) Run(s *flow.Step) {
  if e, ok := s.Arg.(error); ok {
    s.Complete(e)
    return
  }
  e := r.do()
  if e != nil {
    s.Complete(e)
    return
  }
  r.req.bot.op <- &op{what: opNotify}
  s.Complete(nil)
}

func (r *statusNotifyReq) do() error {
  addr, _ := url.Parse(r.req.BaseUrl + statusNotifyUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.req.PassTicket)
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 5)
  m["BaseRequest"] = r.req.BaseReq
  m["Code"] = 3
  m["FromUserName"] = r.req.UserName
  m["ToUserName"] = r.req.UserName
  m["ClientMsgId"] = timestampString13()
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.req.Referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return e
  }
  dumpToFile("5_"+times.NowStrf(times.DateTimeMsFormat5), body)
  return nil
}
