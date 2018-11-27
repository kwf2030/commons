package wechatbot

import (
  "net/http"
  "net/url"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
)

const contactListURL = "/webwxgetcontact"

const ContactListOp = 0x10

type ContactListReq struct {
  req *req
}

func (r *ContactListReq) Run(s *flow.Step) {
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
  data := conv.Slice(resp, "MemberList")
  r.req.op <- &op{What: ContactListOp, Data: data}
  s.Complete(nil)
}

func (r *ContactListReq) checkArg(s *flow.Step) error {
  if e, ok := s.Arg.(error); ok {
    return e
  }
  return nil
}

func (r *ContactListReq) do(s *flow.Step) (map[string]interface{}, error) {
  addr, _ := url.Parse(r.req.baseURL + contactListURL)
  q := addr.Query()
  q.Set("skey", r.req.skey)
  q.Set("pass_ticket", r.req.passTicket)
  q.Set("r", timestampString13())
  q.Set("seq", "0")
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.referer)
  req.Header.Set("User-Agent", userAgent)
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
