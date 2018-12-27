package wechatbot

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "net/url"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
)

const initURL = "/webwxinit"

const ContactSelfOp = 0x01

type InitReq struct {
  req *req
}

func (r *InitReq) Run(s *flow.Step) {
  e := r.checkArg(s)
  if e != nil {
    s.Complete(e)
    return
  }
  resp, e := r.do(s)
  if e != nil {
    s.Complete(e)
    return
  }
  u := conv.GetMap(resp, "User")
  if u == nil {
    s.Complete(ErrInvalidState)
    return
  }
  r.req.userName = conv.GetString(u, "UserName", "")
  r.req.syncKey = conv.GetMap(resp, "SyncKey")
  if r.req.userName == "" || r.req.syncKey == nil {
    s.Complete(ErrInvalidState)
    return
  }
  r.req.avatarURL = fmt.Sprintf("https://%s%s", r.req.host, u["HeadImgUrl"])
  r.req.op <- &op{What: ContactSelfOp, Data: u}
  s.Complete(nil)
}

func (r *InitReq) checkArg(s *flow.Step) error {
  if e, ok := s.Arg.(error); ok {
    return e
  }
  return nil
}

func (r *InitReq) do(s *flow.Step) (map[string]interface{}, error) {
  addr, _ := url.Parse(r.req.baseURL + initURL)
  q := addr.Query()
  q.Set("pass_ticket", r.req.passTicket)
  q.Set("r", timestampString10())
  addr.RawQuery = q.Encode()
  buf, _ := json.Marshal(r.req.payload)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.req.referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  return conv.ReadJSONToMap(resp.Body)
}
