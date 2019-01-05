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

const notifyUrlPath = "/webwxstatusnotify"

const opNotify = 0x5001

// 在手机上显示"已登录Web微信"
type notifyReq struct {
  req *req
}

func (r *notifyReq) Run(s *flow.Step) {
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

func (r *notifyReq) do() error {
  addr, _ := url.Parse(r.req.BaseUrl + notifyUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.req.PassTicket)
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 5)
  m["BaseRequest"] = r.req.BaseReq
  m["ClientMsgId"] = timestampString13()
  m["Code"] = 3
  m["FromUserName"] = r.req.UserName
  m["ToUserName"] = r.req.UserName
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Content-Type", contentType)
  req.Header.Set("Referer", r.req.Referer)
  req.Header.Set("User-Agent", userAgent)
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
