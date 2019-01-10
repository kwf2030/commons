package wxbot

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "sync"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/times"
)

const initUrlPath = "/webwxinit"

const eventInit = 0x4001

type initReq struct {
  *Bot
}

func (r *initReq) Handle(ctx *handlerCtx, evt event) {
  c, e := r.do()
  if e != nil {
    r.syncPipeline.Fire(event{what: eventInit, err: e})
    return
  }
  if c == nil || c.UserName == "" {
    r.syncPipeline.Fire(event{what: eventInit, err: ErrResp})
    return
  }
  sk, ok := c.attr.Load("SyncKey")
  if !ok {
    r.syncPipeline.Fire(event{what: eventInit, err: ErrResp})
    return
  }
  r.session.SyncKey = sk.(syncKey)
  r.session.UserName = c.UserName
  if addr, ok := c.attr.Load("HeadImgUrl"); ok {
    r.session.AvatarUrl = fmt.Sprintf("https://%s%s", r.session.Host, addr.(string))
  }
  r.syncPipeline.Fire(event{what: eventInit, val: c})
  ctx.Fire(evt)
}

func (r *initReq) do() (*Contact, error) {
  addr, _ := url.Parse(r.session.BaseUrl + initUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.session.PassTicket)
  q.Set("r", timestampString10())
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 1)
  m["BaseRequest"] = r.session.BaseReq
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Content-Type", contentType)
  req.Header.Set("Referer", r.session.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  return parseInitResp(resp)
}

func parseInitResp(resp *http.Response) (*Contact, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dump("4_"+times.NowStrf(times.DateTimeMsFormat5), body)
  c := &Contact{raw: body, attr: &sync.Map{}, Friend: &Friend{}, Group: &Group{}}
  jsonparser.EachKey(body, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      sk := parseSyncKey(v)
      if sk.Count > 0 {
        c.attr.Store("SyncKey", sk)
      }
    case 1:
      c.UserName, _ = jsonparser.ParseString(v)
    case 2:
      c.NickName, _ = jsonparser.ParseString(v)
    case 3:
      str, _ := jsonparser.ParseString(v)
      if str != "" {
        c.attr.Store("HeadImgUrl", str)
      }
    }
  }, []string{"SyncKey"}, []string{"User", "UserName"}, []string{"User", "NickName"}, []string{"User", "HeadImgUrl"})
  return c, nil
}
