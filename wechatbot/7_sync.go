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

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const (
  syncCheckUrlPath = "/synccheck"
  syncUrlPath      = "/webwxsync"
)

const (
  opSync       = 0x7001
  opModContact = 0x7002
  opDelContact = 0x7003
  opAddMsg     = 0x7004
  opExit       = 0x7005
)

var (
  jsonPathAddMsgList     = []string{"AddMsgList"}
  jsonPathDelContactList = []string{"DelContactList"}
  jsonPathModContactList = []string{"ModContactList"}
  jsonPathSyncCheckKey   = []string{"SyncCheckKey"}
)

var syncCheckRegex = regexp.MustCompile(`retcode\s*:\s*"(\d+)"\s*,\s*selector\s*:\s*"(\d+)"`)

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
  syncChan := make(chan struct{})
  go r.bridge(ch, syncCheckChan, syncChan)
  go r.syncCheck(ch, syncCheckChan, syncChan)
  go r.sync(ch, syncCheckChan, syncChan)
  syncCheckChan <- struct{}{}
  r.req.bot.op <- &op{what: opSync}
  s.Complete(nil)
}

func (r *syncReq) bridge(ch chan int, syncCheckChan, syncChan chan struct{}) {
  for i := range ch {
    if i == 0 {
      syncCheckChan <- struct{}{}
    } else if i == 1 {
      syncChan <- struct{}{}
    } else {
      close(syncCheckChan)
      close(syncChan)
      break
    }
  }
  close(ch)
}

func (r *syncReq) syncCheck(ch chan int, syncCheckChan, syncChan chan struct{}) {
  for range syncCheckChan {
    code, selector, e := r.doSyncCheck()
    if e != nil {
      times.Sleep()
      ch <- 0
      continue
    }
    if code != 0 {
      ch <- -1
      r.req.bot.op <- &op{what: opExit, syncCheckCode: code, syncCheckSelector: selector}
      close(r.req.bot.op)
      break
    }
    if selector == 0 {
      times.Sleep()
      ch <- 0
      continue
    }
    syncChan <- struct{}{}
  }
}

func (r *syncReq) doSyncCheck() (int, int, error) {
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
    return 0, 0, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return 0, 0, ErrReq
  }
  return parseSyncCheckResp(resp)
}

func (r *syncReq) sync(ch chan int, syncCheckChan, syncChan chan struct{}) {
  for range syncChan {
    data, e := r.doSync()
    if e != nil {
      times.Sleep()
      syncCheckChan <- struct{}{}
      continue
    }
    var addMsgList []*Message
    var modContactList, delContactList []*Contact
    jsonparser.EachKey(data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
      if e != nil {
        return
      }
      switch i {
      case 0:
        addMsgList = parseMsgList(v, r.req.bot)
      case 1:
        delContactList = parseContactList(v, r.req.bot)
      case 2:
        modContactList = parseContactList(v, r.req.bot)
      case 3:
        b, _, _, e := jsonparser.Get(data, "SyncKey")
        if e == nil {
          sk := parseSyncKey(b)
          if sk != nil && sk.Count > 0 {
            r.req.SyncKeys = sk
          }
        }
      }
    }, jsonPathAddMsgList, jsonPathDelContactList, jsonPathModContactList, jsonPathSyncCheckKey)
    // 没开启验证如果被添加好友，
    // ModContactList（对方信息）和AddMsgList（添加到通讯录的系统提示）会一起收到，
    // 所以要先处理完Contact后再处理Message（避免找不到发送者），
    // 虽然之后也能一直收到此人的消息，但要想主动发消息，仍需要手动添加好友，
    // 不添加的话下次登录时好友列表中也没有此人，
    // 目前Web微信好像没有添加好友的功能，所以只能开启验证（通过验证即可添加好友）
    for _, c := range modContactList {
      r.req.bot.op <- &op{what: opModContact, contact: c}
    }
    for _, c := range delContactList {
      r.req.bot.op <- &op{what: opDelContact, contact: c}
    }
    for _, m := range addMsgList {
      r.req.bot.op <- &op{what: opAddMsg, msg: m}
    }
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

func parseSyncCheckResp(resp *http.Response) (int, int, error) {
  // window.synccheck={retcode:"0",selector:"2"}
  // retcode=0：正常，
  // retcode=1100：退出（原因未知），
  // retcode=1101：退出（在手机上点击退出Web微信或长时间没有sychecheck），
  // retcode=1102：退出（原因未知），
  // selector=0：正常，
  // selector=2：有新消息，
  // selector=4：新增或删除联系人/保存群聊到通讯录/修改群名称/群聊成员数目变化，
  // selector=5：未知，
  // selector=6：未知，
  // selector=7：操作了手机（如进入/关闭聊天页面）
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return 0, 0, e
  }
  data := string(body)
  arr := syncCheckRegex.FindStringSubmatch(data)
  if len(arr) < 2 {
    dumpToFile("7_"+times.NowStrf(times.DateTimeMsFormat5)+"_check", body)
    return 0, 0, ErrResp
  }
  code, _ := strconv.Atoi(arr[1])
  selector := 0
  if len(arr) >= 3 {
    selector, _ = strconv.Atoi(arr[2])
  }
  if code != 0 || selector != 0 {
    dumpToFile("7_"+times.NowStrf(times.DateTimeMsFormat5)+"_check", body)
  }
  return code, selector, nil
}

func parseContactList(data []byte, bot *Bot) []*Contact {
  ret := make([]*Contact, 0, 2)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    userName, _ := jsonparser.GetString(v, "UserName")
    if userName == "" {
      return
    }
    c := bot.Contacts.Get(userName)
    if c == nil {
      c = buildContact(v)
      c.withBot(bot)
    }
    ret = append(ret, c)
  })
  return ret
}

func parseMsgList(data []byte, bot *Bot) []*Message {
  ret := make([]*Message, 0, 2)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
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
