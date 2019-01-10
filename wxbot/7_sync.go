package wxbot

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

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/pipeline"
  "github.com/kwf2030/commons/times"
)

const (
  syncCheckUrlPath = "/synccheck"
  syncUrlPath      = "/webwxsync"
)

var syncCheckRegex = regexp.MustCompile(`retcode\s*:\s*"(\d+)"\s*,\s*selector\s*:\s*"(\d+)"`)

var (
  jsonPathAddMsgList     = []string{"AddMsgList"}
  jsonPathDelContactList = []string{"DelContactList"}
  jsonPathModContactList = []string{"ModContactList"}
  jsonPathSyncCheckKey   = []string{"SyncCheckKey"}
)

type syncReq struct {
  *Bot
}

func (r *syncReq) Handle(ctx *pipeline.HandlerContext, val interface{}) {
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
  ctx.Fire(val)
}

func (r *syncReq) bridge(ch chan int, syncCheckChan chan struct{}, syncChan chan syncCheckResp) {
  for i := range ch {
    if i == 0 {
      syncCheckChan <- struct{}{}
    } else if i == 1 {
      syncChan <- syncCheckResp{}
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
    resp, e := r.doSyncCheck()
    if e != nil {
      ch <- 0
      continue
    }
    if resp.code != 0 {
      ch <- -1
      r.Stop()
      r.handler.OnSignOut()
      break
    }
    if resp.selector == 0 {
      ch <- 0
      continue
    }
    syncChan <- resp
  }
}

func (r *syncReq) doSyncCheck() (syncCheckResp, error) {
  addr, _ := url.Parse(fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin%s", r.session.SyncCheckHost, syncCheckUrlPath))
  q := addr.Query()
  q.Set("deviceid", deviceId())
  q.Set("r", timestampString13())
  q.Set("sid", r.session.Sid)
  q.Set("skey", r.session.SKey)
  q.Set("synckey", r.session.SyncKey.expand())
  q.Set("uin", strconv.FormatInt(r.session.Uin, 10))
  q.Set("_", timestampString13())
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.session.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.client.Do(req)
  if e != nil {
    return syncCheckResp{}, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return syncCheckResp{}, ErrReq
  }
  return parseSyncCheckResp(resp)
}

func (r *syncReq) sync(ch chan int, syncCheckChan chan struct{}, syncChan chan syncCheckResp) {
  for syncCheck := range syncChan {
    data, e := r.doSync()
    if e != nil {
      syncCheckChan <- struct{}{}
      continue
    }
    r.dispatch(syncCheck, data)
    syncCheckChan <- struct{}{}
  }
}

func (r *syncReq) doSync() ([]byte, error) {
  addr, _ := url.Parse(r.session.BaseUrl + syncUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.session.PassTicket)
  q.Set("sid", r.session.Sid)
  q.Set("skey", r.session.SKey)
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 3)
  m["BaseRequest"] = r.session.BaseReq
  m["rr"] = strconv.FormatInt(^(times.Timestamp() / int64(time.Second)), 10)
  m["SyncKey"] = r.session.SyncKey
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Content-Type", contentType)
  req.Header.Set("Referer", r.session.Referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.client.Do(req)
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
  dump("7_"+times.NowStrf(times.DateTimeMsFormat5)+"_sync", body)
  return body, nil
}

func (r *syncReq) dispatch(syncCheck syncCheckResp, data []byte) {
  var addMsgList []*Message
  var delContactList, modContactList []*Contact
  jsonparser.EachKey(data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      addMsgList = parseSyncMsgList(v, r.Bot)
    case 1:
      delContactList = parseSyncContactList(v, r.Bot)
    case 2:
      modContactList = parseSyncContactList(v, r.Bot)
    case 3:
      sk := parseSyncKey(v)
      if sk.Count > 0 {
        r.session.SyncKey = sk
      }
    }
  }, jsonPathAddMsgList, jsonPathDelContactList, jsonPathModContactList, jsonPathSyncCheckKey)
  for _, c := range modContactList {
    r.syncPipeline.Fire(c)
  }
  for _, c := range delContactList {
    r.syncPipeline.Fire(c)
  }
  for _, m := range addMsgList {
    r.syncPipeline.Fire(m)
  }
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
    return syncCheckResp{}, e
  }
  data := string(body)
  arr := syncCheckRegex.FindStringSubmatch(data)
  if len(arr) < 2 {
    dump("7_"+times.NowStrf(times.DateTimeMsFormat5)+"_check", body)
    return syncCheckResp{}, ErrResp
  }
  ret := syncCheckResp{}
  ret.code, _ = strconv.Atoi(arr[1])
  if len(arr) >= 3 {
    ret.selector, _ = strconv.Atoi(arr[2])
  }
  if ret.code != 0 || ret.selector != 0 {
    dump("7_"+times.NowStrf(times.DateTimeMsFormat5)+"_check", body)
  }
  return ret, nil
}

func parseSyncContactList(data []byte, bot *Bot) []*Contact {
  ret := make([]*Contact, 0, 2)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    userName, _ := jsonparser.GetString(v, "UserName")
    if userName == "" {
      return
    }
    c := buildContact(v)
    if c != nil && c.UserName != "" {
      c.withBot(bot)
      ret = append(ret, c)
    }
  })
  return ret
}

func parseSyncMsgList(data []byte, bot *Bot) []*Message {
  ret := make([]*Message, 0, 2)
  jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    msg := buildMessage(v)
    if msg != nil && msg.Id != "" {
      msg.withBot(bot)
      ret = append(ret, msg)
    }
  })
  return ret
}

type syncCheckResp struct {
  code     int
  selector int
}
