package wechatbot

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "regexp"

  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const (
  uuidUrl = "https://login.weixin.qq.com/jslogin"
  qrUrl   = "https://login.weixin.qq.com/qrcode"
)

const opQR = 0x1001

var uuidRegex = regexp.MustCompile(`uuid\s*=\s*"(.*)"`)

type uuidReq struct {
  req *req
}

func (r *uuidReq) Run(s *flow.Step) {
  uuid, e := r.do()
  if e != nil {
    s.Complete(e)
    return
  }
  if uuid == "" {
    s.Complete(ErrResp)
    return
  }
  r.req.UUID = uuid
  r.req.QRCodeUrl = fmt.Sprintf("%s/%s", qrUrl, uuid)
  r.req.bot.op <- &op{what: opQR}
  s.Complete(nil)
}

func (r *uuidReq) do() (string, error) {
  addr, _ := url.Parse(uuidUrl)
  q := addr.Query()
  q.Set("appid", "wx782c26e4c19acffb")
  q.Set("fun", "new")
  q.Set("lang", "zh_CN")
  q.Set("_", timestampString13())
  q.Set("redirect_uri", "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage")
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.Referer)
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
  // window.QRLogin.code = 200; window.QRLogin.uuid = "wbVC3cUBrQ==";
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return "", e
  }
  dumpToFile("1_"+times.NowStrf(times.DateTimeMsFormat5), body)
  data := string(body)
  match := uuidRegex.FindStringSubmatch(data)
  if len(match) != 2 {
    return "", ErrResp
  }
  return match[1], nil
}
