package wechatbot

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "regexp"
  "strconv"
  "time"

  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const (
  syncCheckUrlPath = "/synccheck"
  syncUrlPath      = "/webwxsync"
)

const (
  opSync    = 0x7001
  opSignOut = 0x7002
)

var syncCheckRegex = regexp.MustCompile(`retcode\s*:\s*"(\d+)"\s*,\s*selector\s*:\s*"(\d+)"`)

var defaultSyncCheckResp = syncCheckResp{}

type syncReq struct {
  req *req
}

func (r *syncReq) Run(s *flow.Step) {
  if e, ok := s.Arg.(error); ok {
    s.Complete(e)
    return
  }
  // syncCheck一直执行，有消息时才会执行sync，
  // web微信syncCheck的时间间隔约为25秒左右，
  // 即在没有新消息的时候，服务器会保持（阻塞）连接25秒左右
  ch := make(chan int)
  syncCheckChan := make(chan struct{})
  syncChan := make(chan syncCheckResp)
  go r.bridge(ch, syncCheckChan, syncChan)
  go r.syncCheck(ch, syncCheckChan, syncChan)
  go r.sync(ch, syncCheckChan, syncChan)
  syncCheckChan <- struct{}{}
  s.Complete(nil)
}

func (r *syncReq) bridge(ch chan int, syncCheckChan chan struct{}, syncChan chan syncCheckResp) {
  for i := range ch {
    if i == 0 {
      syncCheckChan <- struct{}{}
    } else if i == 1 {
      syncChan <- defaultSyncCheckResp
    } else {
      close(syncCheckChan)
      close(syncChan)
      break
    }
  }
  close(ch)
}

func (r *syncReq) syncCheck(ch chan int, syncCheckChan chan struct{}, syncChan chan syncCheckResp) {
  for range syncCheckChan {
    syncCheck, e := r.doSyncCheck()
    if e != nil {
      times.Sleep()
      ch <- 0
      continue
    }
    if syncCheck.code != 0 {
      ch <- -1
      r.req.bot.op <- op{what: opSignOut, syncCheck: syncCheck}
      close(r.req.bot.op)
      break
    }
    if syncCheck.selector == 0 {
      times.Sleep()
      ch <- 0
      continue
    }
    syncChan <- syncCheck
  }
}

func (r *syncReq) doSyncCheck() (syncCheckResp, error) {
  addr, _ := url.Parse(fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin%s", r.req.SyncCheckHost, syncCheckUrlPath))
  q := addr.Query()
  q.Set("deviceid", deviceId())
  q.Set("r", timestampString13())
  q.Set("sid", r.req.Sid)
  q.Set("skey", r.req.SKey)
  q.Set("synckey", r.req.SyncKeys.expand())
  q.Set("uin", strconv.FormatInt(r.req.Uin, 10))
  q.Set("_", timestampString13())
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return defaultSyncCheckResp, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return defaultSyncCheckResp, ErrReq
  }
  return parseSyncCheckResp(resp)
}

func (r *syncReq) sync(ch chan int, syncCheckChan chan struct{}, syncChan chan syncCheckResp) {
  for syncCheck := range syncChan {
    data, e := r.doSync()
    if e != nil {
      times.Sleep()
      syncCheckChan <- struct{}{}
      continue
    }
    r.req.bot.op <- op{what: opSync, data: data, syncCheck: syncCheck}
    times.Sleep()
    syncCheckChan <- struct{}{}
  }
}

func (r *syncReq) doSync() ([]byte, error) {
  addr, _ := url.Parse(r.req.BaseUrl + syncUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.req.PassTicket)
  q.Set("sid", r.req.Sid)
  q.Set("skey", r.req.SKey)
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 3)
  m["BaseRequest"] = r.req.BaseReq
  m["rr"] = strconv.FormatInt(^(times.Timestamp() / int64(time.Second)), 10)
  m["SyncKey"] = r.req.SyncKeys
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
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("7_"+times.NowStrf(times.DateTimeMsFormat5)+"_sync", body)
  return body, nil
}

func parseSyncCheckResp(resp *http.Response) (syncCheckResp, error) {
  // window.synccheck={retcode:"0",selector:"2"}
  // retcode=0：正常，
  // retcode=1100：退出（原因未知），
  // retcode=1101：退出（在手机上点击退出Web微信或长时间没有synccheck），
  // retcode=1102：退出（原因未知），
  // selector=0：正常，
  // selector=2：有新消息，
  // selector=4：新增或删除联系人/保存群聊到通讯录/修改群名称/群聊成员数目变化，
  // selector=5：未知，
  // selector=6：未知，
  // selector=7：操作了手机（如进入/关闭聊天页面）
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return defaultSyncCheckResp, e
  }
  data := string(body)
  arr := syncCheckRegex.FindStringSubmatch(data)
  if len(arr) < 2 {
    dumpToFile("7_"+times.NowStrf(times.DateTimeMsFormat5)+"_check", body)
    return defaultSyncCheckResp, ErrResp
  }
  ret := syncCheckResp{}
  ret.code, _ = strconv.Atoi(arr[1])
  if len(arr) >= 3 {
    ret.selector, _ = strconv.Atoi(arr[2])
  }
  if ret.code != 0 || ret.selector != 0 {
    dumpToFile("7_"+times.NowStrf(times.DateTimeMsFormat5)+"_check", body)
  }
  return ret, nil
}

type syncCheckResp struct {
  code     int
  selector int
}
