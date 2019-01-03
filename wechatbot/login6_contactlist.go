package wechatbot

import (
  "io/ioutil"
  "net/http"
  "net/url"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/flow"
)

const contactListUrlPath = "/webwxgetcontact"

const opContactList = 0x6001

type contactListReq struct {
  req *req
}

func (r *contactListReq) Run(s *flow.Step) {
  if e, ok := s.Arg.(error); ok {
    s.Complete(e)
    return
  }
  arr, e := r.do()
  if e != nil {
    s.Complete(e)
    return
  }
  r.req.bot.op <- &op{what: opContactList, contacts: arr}
  s.Complete(nil)
}

func (r *contactListReq) do() ([]*Contact, error) {
  addr, _ := url.Parse(r.req.BaseUrl + contactListUrlPath)
  q := addr.Query()
  q.Set("skey", r.req.SKey)
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
  // {
  //   "BaseResponse": {
  //     "Ret": 0,
  //     "ErrMsg": ""
  //   },
  //   "MemberCount": 67,
  //   "MemberList": [
  //     {
  //       "Uin": 0,
  //       "UserName": "weixin",
  //       "NickName": "微信团队",
  //       "HeadImgUrl": "/cgi-bin/mmwebwx-bin/webwxgeticon?seq=123456789&us// ername=weixin&skey=@crypt_123456789abc_123456789abc// ab",
  //       "ContactFlag": 1,
  //       "MemberCount": 0,
  //       "MemberList": [
  //       ],
  //       "RemarkName": "",
  //       "HideInputBarFlag": 0,
  //       "Sex": 0,
  //       "Signature": "微信团队官方帐号",
  //       "VerifyFlag": 56,
  //       "OwnerUin": 0,
  //       "PYInitial": "WXTD",
  //       "PYQuanPin": "weixintuandui",
  //       "RemarkPYInitial": "",
  //       "RemarkPYQuanPin": "",
  //       "StarFriend": 0,
  //       "AppAccountFlag": 0,
  //       "Statues": 0,
  //       "AttrStatus": 4,
  //       "Province": "",
  //       "City": "",
  //       "Alias": "",
  //       "SnsFlag": 0,
  //       "UniFriend": 0,
  //       "DisplayName": "",
  //       "ChatRoomId": 0,
  //       "KeyWord": "wei",
  //       "EncryChatRoomId": "",
  //       "IsOwner": 0
  //     },
  //     {
  //       "Uin": 0,
  //       "UserName": "@123456789abc",
  //       "NickName": "xxx",
  //       "HeadImgUrl": "/cgi-bin/mmwebwx-bin/webwxgeticon?seq=123456789abc&us// ername=@123456789abc&skey=@crypt_123456789abc_123// 123456789abc",
  //       "ContactFlag": 3,
  //       "MemberCount": 0,
  //       "MemberList": [
  //       ],
  //       "RemarkName": "xxx",
  //       "HideInputBarFlag": 0,
  //       "Sex": 2,
  //       "Signature": "xxx",
  //       "VerifyFlag": 0,
  //       "OwnerUin": 0,
  //       "PYInitial": "xxx",
  //       "PYQuanPin": "xxx",
  //       "RemarkPYInitial": "xxx",
  //       "RemarkPYQuanPin": "xxx",
  //       "StarFriend": 0,
  //       "AppAccountFlag": 0,
  //       "Statues": 0,
  //       "AttrStatus": 104503,
  //       "Province": "陕西",
  //       "City": "西安",
  //       "Alias": "",
  //       "SnsFlag": 49,
  //       "UniFriend": 0,
  //       "DisplayName": "",
  //       "ChatRoomId": 0,
  //       "KeyWord": "xxx",
  //       "EncryChatRoomId": "",
  //       "IsOwner": 0
  //     }
  //   ],
  //   "Seq": 0
  // }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
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
