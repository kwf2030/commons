package wechatbot

import (
  "io/ioutil"
  "net/http"
  "net/url"
  "sync"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/flow"
)

const contactListURL = "/webwxgetcontact"

const opContactList = 0x6001

type contactListReq struct {
  req *req
}

func (r *contactListReq) Run(s *flow.Step) {
  arr, e := r.do()
  if e != nil {
    s.Complete(e)
    return
  }
  r.req.op <- &op{what: opContactList, contacts: arr}
  s.Complete(nil)
}

func (r *contactListReq) do() ([]*Contact, error) {
  addr, _ := url.Parse(r.req.BaseUrl + contactListURL)
  q := addr.Query()
  q.Set("skey", r.req.Skey)
  q.Set("pass_ticket", r.req.PassTicket)
  q.Set("r", timestampString13())
  q.Set("seq", "0")
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
  return parseContactListResp(resp)
}

func parseContactListResp(resp *http.Response) ([]*Contact, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  paths := [][]string{{"NickName"}, {"RemarkName"}, {"UserName"}, {"VerifyFlag"}}
  arr := make([]*Contact, 0, 5000)
  _, e = jsonparser.ArrayEach(body, func(v1 []byte, _ jsonparser.ValueType, _ int, e1 error) {
    if e1 != nil {
      return
    }
    c := &Contact{Raw: v1, Attr: &sync.Map{}}
    jsonparser.EachKey(v1, func(i int, v2 []byte, _ jsonparser.ValueType, e2 error) {
      if e2 != nil {
        return
      }
      switch i {
      case 0:
        c.Nickname, _ = jsonparser.ParseString(v2)
      case 1:
        c.RemarkName, _ = jsonparser.ParseString(v2)
      case 2:
        c.UserName, _ = jsonparser.ParseString(v2)
      case 3:
        flag, _ := jsonparser.ParseInt(v2)
        if flag != 0 {
          c.VerifyFlag = int(flag)
        }
      }
    }, paths...)
    if c.UserName != "" {
      arr = append(arr, c)
    }
  }, "MemberList")
  if e != nil {
    return nil, e
  }
  return arr, nil
}
