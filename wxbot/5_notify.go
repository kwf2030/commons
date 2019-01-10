package wxbot

import (
  "bytes"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "net/url"

  "github.com/kwf2030/commons/times"
)

const notifyUrlPath = "/webwxstatusnotify"

const eventNotify = 0x5001

type notifyReq struct {
  *Bot
}

func (r *notifyReq) Handle(ctx *handlerCtx, evt event) {
  e := r.do()
  if e != nil {
    r.syncPipeline.Fire(event{what: eventNotify, err: e})
    return
  }
  r.syncPipeline.Fire(event{what: eventNotify})
  ctx.Fire(evt)
}

func (r *notifyReq) do() error {
  addr, _ := url.Parse(r.session.BaseUrl + notifyUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.session.PassTicket)
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 5)
  m["BaseRequest"] = r.session.BaseReq
  m["ClientMsgId"] = timestampString13()
  m["Code"] = 3
  m["FromUserName"] = r.session.UserName
  m["ToUserName"] = r.session.UserName
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Content-Type", contentType)
  req.Header.Set("Referer", r.session.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.client.Do(req)
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
  dump("5_"+times.NowStrf(times.DateTimeMsFormat5), body)
  return nil
}
