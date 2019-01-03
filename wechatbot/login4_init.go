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
  jsonPathSyncKey        = []string{"SyncKey"}
  jsonPathUserHeadImgUrl = []string{"User", "HeadImgUrl"}
  jsonPathUserNickName   = []string{"User", "NickName"}
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
  // {
  //   "BaseResponse": {
  //     "Ret": 0,
  //     "ErrMsg": ""
  //   },
  //   "Count": 2,
  //   "ContactList": [
  //     {
  //       "Uin": 0,
  //       "UserName": "filehelper",
  //       "NickName": "文件传输助手",
  //       ...
  //     },
  //     {
  //       "Uin": 0,
  //       "UserName": "weixin",
  //       "NickName": "微信团队",
  //       ...
  //     }
  //   ],
  //   "SyncKey": {
  //     "Count": 4,
  //     "List": [
  //       {
  //         "Key": 1,
  //         "Val": 123456789
  //       },
  //       ...
  //     ]
  //   },
  //   "User": {
  //     "Uin": 123456,
  //     "UserName": "@ xxx//xxx",
  //     "NickName": "xxx",
  //     "HeadImgUrl": "/cgi-bin/mmwebwx-bin/webwxgeticon?seq=123456789&user //name=@123456789// 123&skey=@crypt_123456789abc_123456789abc",
  //     "RemarkName": "",
  //     "PYInitial": "",
  //     "PYQuanPin": "",
  //     "RemarkPYInitial": "",
  //     "RemarkPYQuanPin": "",
  //     "HideInputBarFlag": 0,
  //     "StarFriend": 0,
  //     "Sex": 0,
  //     "Signature": "xxx",
  //     "AppAccountFlag": 0,
  //     "VerifyFlag": 0,
  //     "ContactFlag": 0,
  //     "WebWxPluginSwitch": 0,
  //     "HeadImgFlag": 1,
  //     "SnsFlag": 0
  //   },
  //   "ChatSet": "filehelper,weixin,",
  //   "SKey": "@crypt_123456789abc_123456789abc",
  //   "ClientVersion": 123456789,
  //   "SystemTime": 123456789,
  //   "GrayScale": 1,
  //   "InviteStartCount": 40,
  //   "MPSubscribeMsgCount": 0,
  //   "MPSubscribeMsgList": [
  //   ],
  //   "ClickReportInterval": 600000
  // }
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
      sk := parseSyncKey(v)
      if sk != nil && sk.Count > 0 {
        c.Attr.Store("SyncKeys", sk)
      }
    case 1:
      str, _ := jsonparser.ParseString(v)
      if str != "" {
        c.Attr.Store("HeadImgUrl", str)
      }
    case 2:
      c.NickName, _ = jsonparser.ParseString(v)
    case 3:
      c.UserName, _ = jsonparser.ParseString(v)
    }
  }, jsonPathSyncKey, jsonPathUserHeadImgUrl, jsonPathUserNickName, jsonPathUserUserName)
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
