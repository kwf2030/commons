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

const initUrlPath = "/webwxinit"

const opInit = 0x4001

var (
  jsonPathUserHeadImgUrl = []string{"User", "HeadImgUrl"}
  jsonPathUserNickName   = []string{"User", "NickName"}
  jsonPathUserSyncKey    = []string{"User", "SyncKey"}
  jsonPathUserUserName   = []string{"User", "UserName"}
)

type initReq struct {
  req *req
}

func (r *initReq) Run(s *flow.Step) {
  if e, ok := s.Arg.(error); ok {
    s.Complete(e)
    return
  }
  c, e := r.do()
  if e != nil {
    s.Complete(e)
    return
  }
  if c == nil || c.UserName == "" {
    s.Complete(ErrResp)
    return
  }
  sk, ok := c.Attr.Load("SyncKeys")
  if !ok {
    s.Complete(ErrResp)
    return
  }
  if addr, ok := c.Attr.Load("HeadImgUrl"); ok {
    r.req.AvatarUrl = fmt.Sprintf("https://%s%s", r.req.Host, addr.(string))
    c.Attr.Delete("HeadImgUrl")
  }
  r.req.SyncKeys = sk.(*syncKeys)
  c.Attr.Delete("SyncKeys")
  r.req.UserName = c.UserName
  // todo 给Bot.Self赋值，Bot.Self需要的session信息保存到Contact.Attr中（Contact没有对应字段的话），Self.Id需要Contacts初始化完成之后才能计算出来
  r.req.bot.op <- &op{what: opInit, contact: c}
  s.Complete(nil)
}

func (r *initReq) do() (*Contact, error) {
  addr, _ := url.Parse(r.req.BaseUrl + initUrlPath)
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
  c := &Contact{Raw: body, Attr: &sync.Map{}}
  jsonparser.EachKey(body, func(i int, v []byte, _ jsonparser.ValueType, e error) {
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
      c.NickName, _ = jsonparser.ParseString(v)
    case 2:
      sk := &syncKeys{}
      json.Unmarshal(v, sk)
      if sk.Count > 0 {
        c.Attr.Store("SyncKeys", sk)
      }
    case 3:
      c.UserName, _ = jsonparser.ParseString(v)
    }
  }, jsonPathUserHeadImgUrl, jsonPathUserNickName, jsonPathUserSyncKey, jsonPathUserUserName)
  return c, nil
}

type syncKey struct {
  Key int `json:"Key"`
  Val int `json:"Val"`
}

type syncKeys struct {
  Count int        `json:"Count"`
  List  []*syncKey `json:"List"`
}

func (sk *syncKeys) expand() string {
  var e error
  var sb strings.Builder
  n := sk.Count - 1
  for i := 0; i <= n; i++ {
    item := sk.List[i]
    _, e = fmt.Fprintf(&sb, "%d_%d", item.Key, item.Val)
    if e != nil {
      return ""
    }
    if i != n {
      sb.WriteString("|")
    }
  }
  return sb.String()
}
