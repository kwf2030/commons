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

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const (
  syncCheckURL = "/synccheck"
  syncURL      = "/webwxsync"
)

const (
  opSync       = 0x6001
  opMsg        = 0x6002
  opContactMod = 0x6003
  opContactDel = 0x6004
  opExit       = 0x6005
)

var syncCheckRegexp = regexp.MustCompile(`retcode\s*:\s*"(\d+)"\s*,\s*selector\s*:\s*"(\d+)"`)

type syncReq struct {
  req *req
}

func (r *syncReq) Run(s *flow.Step) {
  // syncCheck一直执行，有消息时才会执行sync，
  // web微信syncCheck的时间间隔约为25秒左右，
  // 即在没有新消息的时候，服务器会保持（阻塞）连接25秒左右
  ch1 := make(chan struct{})
  ch2 := make(chan struct{})
  go r.syncCheck(ch1, ch2)
  go r.sync(ch1, ch2)
  ch1 <- struct{}{}
  r.req.op <- &op{what: opSync}
  s.Complete(nil)
}

func (r *syncReq) syncCheck(ch1 chan struct{}, ch2 chan struct{}) {
  for range ch1 {
    code, selector, e := r.doSyncCheck()
    switch {
    case e != nil:
      fallthrough

    case code == 0 && selector == 0:
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      ch1 <- struct{}{}

    case code != 0:
      close(ch1)
      close(ch2)
      r.req.op <- &op{what: TerminateOp, Data: code}
      close(r.req.op)

    default:
      ch2 <- struct{}{}
    }
  }
}

func (r *syncReq) sync(ch1 chan struct{}, ch2 chan struct{}) {
  for range ch2 {
    resp, e := r.doSync()
    switch {
    case e != nil, resp == nil:
      fallthrough

    case conv.GetInt(conv.GetMap(resp, "BaseResponse"), "Ret", 0) != 0:
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      ch1 <- struct{}{}
      continue
    }

    r.req.SyncKeys = conv.GetMap(resp, "SyncCheckKey")

    // 没开启验证如果被添加好友，
    // ModContactList（对方信息）和AddMsgList（添加到通讯录的系统提示）会一起收到，
    // 要先处理完Contact后再处理Message（否则会出现找不到发送者的问题），
    // 虽然之后也能一直收到此人的消息，但要想主动发消息，仍需要手动添加好友，
    // 不添加的话下次登录时好友列表中也没有此人，
    // 目前Web微信好像没有添加好友的功能，所以只能开启验证（通过验证即可添加好友）
    if conv.GetInt(resp, "ModContactCount", 0) > 0 {
      data := conv.GetMapSlice(resp, "ModContactList")
      for _, v := range data {
        r.req.op <- &op{what: ContactModOp, Data: v}
      }
    }
    if conv.GetInt(resp, "DelContactCount", 0) > 0 {
      data := conv.GetMapSlice(resp, "DelContactList")
      for _, v := range data {
        r.req.op <- &op{what: ContactDelOp, Data: v}
      }
    }
    if conv.GetInt(resp, "AddMsgCount", 0) > 0 {
      data := conv.GetMapSlice(resp, "AddMsgList")
      for _, v := range data {
        r.req.op <- &op{what: MsgOp, Data: v}
      }
    }

    time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
    ch1 <- struct{}{}
  }
}

// 检查是否有新消息，类似于心跳，
// window.synccheck={retcode:"0",selector:"2"}
// retcode=0：正常，
// retcode=1100：失败/已退出，
// retcode=1101：在其他地方登录了Web微信，
// retcode=1102：主动退出，
// selector=0：正常，
// selector=2：有新消息，
// selector=4：保存群聊到通讯录/修改群名称/新增或删除联系人/群聊成员数目变化，
// selector=5：未知，
// selector=6：未知，
// selector=7：操作了手机，如进入/关闭聊天页面
func (r *syncReq) doSyncCheck() (int, int, error) {
  addr, _ := url.Parse(fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin%s", r.req.SyncCheckHost, syncCheckURL))
  q := addr.Query()
  q.Set("r", timestampString13())
  q.Set("sid", r.req.Sid)
  q.Set("uin", strconv.Itoa(r.req.Uin))
  q.Set("skey", r.req.Skey)
  q.Set("deviceid", deviceID())
  q.Set("synckey", r.req.SyncKeys.flat())
  q.Set("_", timestampString13())
  addr.RawQuery = q.Encode()
  // 请求必须加上Cookies
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return 0, 0, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return 0, 0, ErrReq
  }
  return parseSyncCheckResp(resp)
}

func (r *syncReq) doSync() (map[string]interface{}, error) {
  addr, _ := url.Parse(r.req.BaseUrl + syncURL)
  q := addr.Query()
  q.Set("pass_ticket", r.req.PassTicket)
  q.Set("sid", r.req.Sid)
  q.Set("skey", r.req.Skey)
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 3)
  m["BaseRequest"] = r.req.BaseReq
  m["SyncKey"] = r.req.SyncKeys
  m["rr"] = strconv.FormatInt(^(times.Timestamp() / int64(time.Second)), 10)
  buf, _ := json.Marshal(m)
  // 请求必须加上Content-Type和Cookies
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
  return conv.ReadJSONToMap(resp.Body)
}

func parseSyncCheckResp(resp *http.Response) (int, int, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return 0, 0, e
  }
  data := string(body)
  match := syncCheckRegexp.FindStringSubmatch(data)
  if len(match) < 2 {
    return 0, 0, ErrResp
  }
  code, _ := strconv.Atoi(match[1])
  selector := 0
  if len(match) >= 3 {
    selector, _ = strconv.Atoi(match[2])
  }
  return code, selector, nil
}
