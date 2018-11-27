package wechatbot

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "regexp"

  "github.com/kwf2030/commons/flow"
)

const (
  uuidURL = "https://login.weixin.qq.com/jslogin"
  qrURL   = "https://login.weixin.qq.com/qrcode"
)

var uuidRegexp = regexp.MustCompile(`uuid\s*=\s*"(.*)"`)

type UUIDReq struct {
  req *req
}

func (r *UUIDReq) Run(s *flow.Step) {
  e := r.checkArg(s)
  if e != nil {
    s.Complete(e)
    return
  }
  uuid, e := r.do(s)
  if e != nil {
    s.Complete(e)
    return
  }
  if uuid == "" {
    s.Complete(ErrInvalidState)
    return
  }
  r.req.uuid = uuid
  qrChan := s.Arg.(chan<- string)
  qrChan <- fmt.Sprintf("%s/%s", qrURL, uuid)
  close(qrChan)
  s.Complete(nil)
}

func (r *UUIDReq) checkArg(s *flow.Step) error {
  if s.Arg == nil {
    return ErrInvalidArgs
  }
  if e, ok := s.Arg.(error); ok {
    return e
  }
  return nil
}

func (r *UUIDReq) do(s *flow.Step) (string, error) {
  addr, _ := url.Parse(uuidURL)
  q := addr.Query()
  q.Set("appid", "wx782c26e4c19acffb")
  q.Set("fun", "new")
  q.Set("lang", "zh_CN")
  q.Set("_", timestampString13())
  q.Set("redirect_uri", "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage")
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return "", ErrReq
  }
  return parseUUIDResp(resp)
}

func parseUUIDResp(resp *http.Response) (string, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return "", e
  }
  data := string(body)
  match := uuidRegexp.FindStringSubmatch(data)
  if len(match) != 2 {
    return "", ErrResp
  }
  return match[1], nil
}
