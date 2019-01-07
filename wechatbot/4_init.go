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
  "github.com/kwf2030/commons/times"
)

const initUrlPath = "/webwxinit"

const opInit = 0x4001

var (
  jsonPathSyncKey        = []string{"SyncKey"}
  jsonPathUserUserName   = []string{"User", "UserName"}
  jsonPathUserNickName   = []string{"User", "NickName"}
  jsonPathUserHeadImgUrl = []string{"User", "HeadImgUrl"}
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
  r.req.UserName = c.UserName
  r.req.SyncKeys = sk.(*syncKeys)
  c.Attr.Delete("SyncKeys")
  if addr, ok := c.Attr.Load("HeadImgUrl"); ok {
    r.req.AvatarUrl = fmt.Sprintf("https://%s%s", r.req.Host, addr.(string))
    c.Attr.Delete("HeadImgUrl")
  }
  r.req.bot.op <- op{what: opInit, data: c}
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
  req.Header.Set("Content-Type", contentType)
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
  return parseInitResp(resp)
}

func parseInitResp(resp *http.Response) (*Contact, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("4_"+times.NowStrf(times.DateTimeMsFormat5), body)
  c := &Contact{Raw: body, Attr: &sync.Map{}, Friend: &Friend{}, Group: &Group{}}
  jsonparser.EachKey(body, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      sk := parseSyncKey(v)
      if sk != nil && sk.Count > 0 {
        c.Attr.Store("SyncKeys", sk)
      }
    case 1:
      c.UserName, _ = jsonparser.ParseString(v)
    case 2:
      c.NickName, _ = jsonparser.ParseString(v)
    case 3:
      str, _ := jsonparser.ParseString(v)
      if str != "" {
        c.Attr.Store("HeadImgUrl", str)
      }
    }
  }, jsonPathSyncKey, jsonPathUserUserName, jsonPathUserNickName, jsonPathUserHeadImgUrl)
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

func parseSyncKey(data []byte) *syncKeys {
  n, _ := jsonparser.GetInt(data, "Count")
  if n <= 0 {
    return nil
  }
  ret := &syncKeys{Count: int(n), List: make([]*syncKey, 0, n)}
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, i int, e error) {
    if e != nil {
      return
    }
    key, _ := jsonparser.GetInt(v, "Key")
    val, _ := jsonparser.GetInt(v, "Val")
    ret.List = append(ret.List, &syncKey{int(key), int(val)})
  }, "List")
  return ret
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
