package wechatbot

import (
  "encoding/xml"
  "io/ioutil"
  "net/http"
  "net/url"
  "strings"

  "github.com/kwf2030/commons/flow"
)

const opLogin = 0x3001

type loginReq struct {
  req *req
}

func (r *loginReq) Run(s *flow.Step) {
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
  r.req.bot.op <- &op{what: opLogin}
  s.Complete(nil)
}

func (r *loginReq) do() (*loginResp, error) {
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
  return parseLoginResp(resp)
}

func (r *loginReq) selectBaseUrl() {
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

func parseLoginResp(resp *http.Response) (*loginResp, error) {
  // <error>
  //   <ret>0</ret>
  //   <message></message>
  //   <skey>@crypt_7be66f24_74523cd7d1f1f5cdb34b1fb1ea5040cf</skey>
  //   <wxsid>cbj6tfPFvr+Eq08O</wxsid>
  //   <wxuin>951790778</wxuin>
  //   <pass_ticket>xvDQbSZVALSZQ%2BiCR%2BGI83zvZwLoGOKViqY2qeJmg2fgSikwR85a7ipplVbOYeuv</pass_ticket>
  //   <isgrayscale>1</isgrayscale>
  // </error>
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  ret := &loginResp{}
  e = xml.Unmarshal(body, ret)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

type loginResp struct {
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
