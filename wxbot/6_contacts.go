package wxbot

import (
  "io/ioutil"
  "net/http"
  "net/url"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/times"
)

const contactsUrlPath = "/webwxgetcontact"

const eventContacts = 0x6001

type contactsReq struct {
  *Bot
}

func (r *contactsReq) Handle(ctx *handlerCtx, evt event) {
  arr, e := r.do()
  if e != nil {
    r.syncPipeline.Fire(event{what: eventContacts, err: e})
    return
  }
  r.syncPipeline.Fire(event{what: eventContacts, val: arr})
  ctx.Fire(evt)
}

func (r *contactsReq) do() ([]*Contact, error) {
  addr, _ := url.Parse(r.session.BaseUrl + contactsUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.session.PassTicket)
  q.Set("r", timestampString13())
  q.Set("seq", "0")
  q.Set("skey", r.session.SKey)
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
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
  return parseContactsResp(resp)
}

func parseContactsResp(resp *http.Response) ([]*Contact, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dump("6_"+times.NowStrf(times.DateTimeMsFormat5), body)
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
