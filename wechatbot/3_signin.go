package wechatbot

import (
  "encoding/xml"
  "io/ioutil"
  "net/http"
  "net/url"
  "strings"

  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const opSignIn = 0x3001

type signInReq struct {
  req *req
}

func (r *signInReq) Run(s *flow.Step) {
  if e, ok := s.Arg.(error); ok {
    s.Complete(e)
    return
  }
  login, e := r.do()
  if e != nil {
    s.Complete(e)
    return
  }
  if login == nil || login.WXUin == 0 || login.WXSid == "" || login.SKey == "" || login.PassTicket == "" {
    s.Complete(ErrResp)
    return
  }
  r.req.SKey = login.SKey
  r.req.Sid = login.WXSid
  r.req.Uin = login.WXUin
  r.req.PassTicket = login.PassTicket
  r.req.BaseReq = &baseReq{
    SKey:     login.SKey,
    Sid:      login.WXSid,
    Uin:      login.WXUin,
    DeviceId: deviceId(),
  }
  r.selectBaseUrl()
  r.req.bot.op <- &op{what: opSignIn}
  s.Complete(nil)
}

func (r *signInReq) do() (*signInResp, error) {
  u, _ := url.Parse(r.req.RedirectUrl)
  // 返回的地址可能没有fun和version两个参数，而此请求必须这两个参数
  q := u.Query()
  q.Set("fun", "new")
  q.Set("version", "v2")
  u.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", u.String(), nil)
  req.Header.Set("Referer", r.req.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrResp
  }
  return parseSignInResp(resp)
}

func (r *signInReq) selectBaseUrl() {
  u, _ := url.Parse(r.req.RedirectUrl)
  host := u.Hostname()
  r.req.Host = host
  switch {
  case strings.Contains(host, "wx2"):
    r.req.SyncCheckHost = "webpush.wx2.qq.com"
    r.req.Host = "wx2.qq.com"
    r.req.Referer = "https://wx2.qq.com/"
    r.req.BaseUrl = "https://wx2.qq.com/cgi-bin/mmwebwx-bin"
  }
}

func parseSignInResp(resp *http.Response) (*signInResp, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("3_"+times.NowStrf(times.DateTimeMsFormat5), body)
  ret := &signInResp{}
  e = xml.Unmarshal(body, ret)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

type signInResp struct {
  XMLName     xml.Name `xml:"error"`
  Ret         int      `xml:"ret"`
  Message     string   `xml:"message"`
  WXUin       int64    `xml:"wxuin"`
  WXSid       string   `xml:"wxsid"`
  SKey        string   `xml:"skey"`
  PassTicket  string   `xml:"pass_ticket"`
  IsGrayScale int      `xml:"isgrayscale"`
}

type baseReq struct {
  SKey     string `json:"Skey"`
  Sid      string `json:"Sid"`
  Uin      int64  `json:"Uin"`
  DeviceId string `json:"DeviceID"`
}