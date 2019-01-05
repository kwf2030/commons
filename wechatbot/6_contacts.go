package wechatbot

import (
  "io/ioutil"
  "net/http"
  "net/url"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const contactsUrlPath = "/webwxgetcontact"

const opContacts = 0x6001

type contactsReq struct {
  req *req
}

func (r *contactsReq) Run(s *flow.Step) {
  if e, ok := s.Arg.(error); ok {
    s.Complete(e)
    return
  }
  arr, e := r.do()
  if e != nil {
    s.Complete(e)
    return
  }
  r.req.bot.op <- &op{what: opContacts, contacts: arr}
  s.Complete(nil)
}

func (r *contactsReq) do() ([]*Contact, error) {
  addr, _ := url.Parse(r.req.BaseUrl + contactsUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.req.PassTicket)
  q.Set("r", timestampString13())
  q.Set("seq", "0")
  q.Set("skey", r.req.SKey)
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  return parseContactsResp(resp)
}

func parseContactsResp(resp *http.Response) ([]*Contact, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("6_"+times.NowStrf(times.DateTimeMsFormat5), body)
  arr := make([]*Contact, 0, 5000)
  _, e = jsonparser.ArrayEach(body, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    c := buildContact(v)
    if c != nil && c.UserName != "" {
      arr = append(arr, c)
    }
  }, "MemberList")
  if e == nil || e == jsonparser.KeyPathNotFoundError {
    return arr, nil
  }
  return nil, e
}
