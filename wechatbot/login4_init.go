package wechatbot

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "strings"
  "sync"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/flow"
)

const initURL = "/webwxinit"

const opInit = 0x4001

type initReq struct {
  req *req
}

func (r *initReq) Run(s *flow.Step) {
  c, e := r.do()
  if e != nil {
    s.Complete(e)
    return
  }
  if c == nil || c.UserName == "" {
    s.Complete(ErrResp)
    return
  }
  sk, ok := c.Attr.Load("SyncKey")
  if !ok {
    s.Complete(ErrResp)
    return
  }
  r.req.SyncKeys = sk.(*syncKeys)
  c.Attr.Delete("SyncKey")
  r.req.UserName = c.UserName
  if addr, ok := c.Attr.Load("HeadImgUrl"); ok {
    r.req.AvatarURL = fmt.Sprintf("https://%s%s", r.req.Host, addr.(string))
    c.Attr.Delete("HeadImgUrl")
  }
  r.req.op <- &op{what: opInit, contact: c}
  s.Complete(nil)
}

func (r *initReq) do() (*Contact, error) {
  addr, _ := url.Parse(r.req.BaseUrl + initURL)
  q := addr.Query()
  q.Set("pass_ticket", r.req.PassTicket)
  q.Set("r", timestampString10())
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 1)
  m["BaseRequest"] = r.req.BaseReq
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.req.Referer)
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
  return parseInitResp(resp)
}

func parseInitResp(resp *http.Response) (*Contact, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  paths := [][]string{{"User", "HeadImgUrl"}, {"User", "NickName"}, {"User", "SyncKey"}, {"User", "UserName"}}
  c := &Contact{Raw: body, Attr: &sync.Map{}}
  jsonparser.EachKey(body, func(i int, v []byte, t jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      str, _ := jsonparser.ParseString(v)
      if str != "" {
        c.Attr.Store("HeadImgUrl", str)
      }
    case 1:
      c.Nickname, _ = jsonparser.ParseString(v)
    case 2:
      sk := &syncKeys{}
      e = json.Unmarshal(v, sk)
      if e != nil {
        return
      }
      if sk.Count > 0 {
        c.Attr.Store("SyncKey", sk)
      }
    case 3:
      c.UserName, _ = jsonparser.ParseString(v)
    }
  }, paths...)
  return c, nil
}

type syncKey struct {
  Key int
  Val int
}

type syncKeys struct {
  Count int
  List  []*syncKey
}

func (sk *syncKeys) flat() string {
  var e error
  var sb strings.Builder
  for i := 0; i < sk.Count; i++ {
    _, e = fmt.Fprintf(&sb, "%d_%d", sk.List[i].Key, sk.List[i].Val)
    if e != nil {
      return ""
    }
    if i != sk.Count-1 {
      sb.WriteString("|")
    }
  }
  return sb.String()
}
