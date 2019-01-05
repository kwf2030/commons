package wechatbot

import (
  "io/ioutil"
  "net/http"
  "net/url"
  "regexp"
  "strconv"
  "time"

  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const scanUrl = "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login"

const opScan = 0x2001

var (
  scanStCodeRegex        = regexp.MustCompile(`code\s*=\s*(\d+)\s*;`)
  scanStRedirectURLRegex = regexp.MustCompile(`redirect_uri\s*=\s*"(.*)"`)
)

type scanReq struct {
  req *req
}

func (r *scanReq) Run(s *flow.Step) {
  if e, ok := s.Arg.(error); ok {
    s.Complete(e)
    return
  }
  ch := make(chan string)
  go r.check(ch)
  redirectUrl := <-ch
  close(ch)
  if redirectUrl == "" {
    // 如果是空，基本就是超时（一直没有扫描默认设置了2分钟超时），
    // 微信基本不可能返回200状态码的同时返回空redirect_url
    s.Complete(ErrTimeout)
    return
  }
  r.req.RedirectUrl = redirectUrl
  r.req.bot.op <- &op{what: opScan}
  s.Complete(nil)
}

func (r *scanReq) check(ch chan<- string) {
  loop := true
  t := time.AfterFunc(time.Minute*2, func() {
    if loop {
      loop = false
      r.req.State = StateTimeout
      ch <- ""
    }
  })
out:
  for loop {
    // 200（已确认），201（已扫描），408（未扫描）
    code, addr, e := r.do()
    if e != nil {
      times.Sleep()
      continue
    }
    switch code {
    case 200:
      t.Stop()
      loop = false
      r.req.State = StateConfirmed
      ch <- addr
      break out

    case 201:
      r.req.State = StateScanned
      times.Sleep()
      continue

    case 408:
      t.Stop()
      loop = false
      r.req.State = StateTimeout
      ch <- ""
      break out
    }
  }
}

func (r *scanReq) do() (int, string, error) {
  addr, _ := url.Parse(scanUrl)
  q := addr.Query()
  q.Set("loginicon", "true")
  q.Set("r", timestampString10())
  q.Set("tip", "0")
  q.Set("uuid", r.req.UUID)
  q.Set("_", timestampString13())
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return 0, "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return 0, "", ErrReq
  }
  // RedirectURL的Host可能是wx.qq.com、wx2.qq.com或其他地址，
  // 这个地址可能是根据帐号注册时间分配的，
  // 从下一步reqToken开始所有的请求必须使用相同的Host，否则会返回1100错误码，
  // wx2版本有些请求的query参数被省略了，暂时不用管
  return parseScanResp(resp)
}

func parseScanResp(resp *http.Response) (int, string, error) {
  // 如果是200，返回：window.code=200;window.redirect_uri=xxx
  // 如果是201，返回：window.code=201;window.userAvatar = 'data:img/jpg;base64,xxx'
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return 0, "", e
  }
  dumpToFile("2_"+times.NowStrf(times.DateTimeMsFormat5), body)
  data := string(body)
  arr := scanStCodeRegex.FindStringSubmatch(data)
  if len(arr) != 2 {
    return 0, "", ErrResp
  }
  code, e := strconv.Atoi(arr[1])
  if e != nil {
    return 0, "", ErrResp
  }
  if code != 200 {
    return code, "", nil
  }
  arr = scanStRedirectURLRegex.FindStringSubmatch(data)
  if len(arr) < 2 {
    return code, "", ErrResp
  }
  return code, arr[1], nil
}
