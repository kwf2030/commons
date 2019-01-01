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
  jsonPathSyncCheckKey   = []string{"SyncCheckKey"}
  jsonPathModContactList = []string{"ModContactList"}
  jsonPathDelContactList = []string{"DelContactList"}
  jsonPathAddMsgList     = []string{"AddMsgList"}
)

var syncCheckRegex = regexp.MustCompile(`retcode\s*:\s*"(\d+)"\s*,\s*selector\s*:\s*"(\d+)"`)

type syncReq struct {
  req *req
}

func (r *syncReq) Run(s *flow.Step) {
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
  r.req.op <- &op{what: opSync}
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
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      ch <- 0
      continue
    }
    if code != 0 {
      ch <- -1
      r.req.op <- &op{what: opExit, syncCheckCode: code, syncCheckSelector: selector}
      close(r.req.op)
      break
    }
    if selector == 0 {
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      ch <- 0
      continue
    }
    syncChan <- struct{}{}
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
  addr, _ := url.Parse(fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin%s", r.req.SyncCheckHost, syncCheckUrlPath))
  q := addr.Query()
  q.Set("r", timestampString13())
  q.Set("sid", r.req.Sid)
  q.Set("uin", strconv.FormatInt(r.req.Uin, 10))
  q.Set("skey", r.req.SKey)
  q.Set("deviceid", deviceId())
  q.Set("synckey", r.req.SyncKeys.expand())
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
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      syncCheckChan <- struct{}{}
      continue
    }
    var modContactList, delContactList []*Contact
    var addMsgList []*Message
    jsonparser.EachKey(data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
      if e != nil {
        return
      }
      switch i {
      case 0:
        b, _, _, e := jsonparser.Get(data, "SyncKey")
        if e == nil {
          sk := parseSyncKey(b)
          if sk != nil && sk.Count > 0 {
            r.req.SyncKeys = sk
          }
        }
      case 1:
        modContactList = parseModContact(v)
      case 2:
        delContactList = parseDelContact(v)
      case 3:
        addMsgList = parseAddMsg(v)
      }
    }, jsonPathSyncCheckKey, jsonPathModContactList, jsonPathDelContactList, jsonPathAddMsgList)
    for _, c := range modContactList {
      r.req.op <- &op{what: opModContact, contact: c}
    }
    for _, c := range delContactList {
      r.req.op <- &op{what: opDelContact, contact: c}
    }
    for _, m := range addMsgList {
      r.req.op <- &op{what: opAddMsg, msg: m}
    }
    time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
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
  m["SyncKey"] = r.req.SyncKeys
  m["rr"] = strconv.FormatInt(^(times.Timestamp() / int64(time.Second)), 10)
  buf, _ := json.Marshal(m)
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
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  return body, nil
}

func parseSyncCheckResp(resp *http.Response) (int, int, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return 0, 0, e
  }
  data := string(body)
  arr := syncCheckRegex.FindStringSubmatch(data)
  if len(arr) < 2 {
    return 0, 0, ErrResp
  }
  code, _ := strconv.Atoi(arr[1])
  selector := 0
  if len(arr) >= 3 {
    selector, _ = strconv.Atoi(arr[2])
  }
  return code, selector, nil
}

func parseSyncKey(data []byte) *syncKeys {
  ret := &syncKeys{List: make([]*syncKey, 0, 11)}
  n, _ := jsonparser.GetInt(data, "Count")
  ret.Count = int(n)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, i int, e error) {
    if e != nil {
      return
    }
    key, _ := jsonparser.GetInt(v, "Key")
    val, _ := jsonparser.GetInt(v, "Val")
    ret.List = append(ret.List, &syncKey{int(key), int(val)})
  }, "List")
  return ret
}

func parseModContact(data []byte) []*Contact {
  /*arr := make([]*Contact, 0, count)
   _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
     if e != nil {
       return
     }
     c := buildContact(v)
     if c != nil && c.UserName != "" {
       arr = append(arr, c)
     }
   }, jsonKeyModContactList)*/
  // todo 通知
  return nil
}

func parseDelContact(data []byte) []*Contact {
  /*arr := make([]*Contact, 0, count)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    c := buildContact(v)
    if c != nil && c.UserName != "" {
      arr = append(arr, c)
    }
  }, jsonKeyDelContactList)*/
  // todo 通知
  return nil
}

func parseAddMsg(data []byte) []*Message {
  /*arr := make([]*Message, 0, count)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    msg := buildMessage(v)
    if msg != nil && msg.Id != "" {
      arr = append(arr, msg)
    }
  }, jsonKeyAddMsgList)*/
  // todo 通知
  // 没开启验证如果被添加好友，
  // ModContactList（对方信息）和AddMsgList（添加到通讯录的系统提示）会一起收到，
  // 要先处理完Contact后再处理Message（否则会出现找不到发送者的问题），
  // 虽然之后也能一直收到此人的消息，但要想主动发消息，仍需要手动添加好友，
  // 不添加的话下次登录时好友列表中也没有此人，
  // 目前Web微信好像没有添加好友的功能，所以只能开启验证（通过验证即可添加好友）
  return nil
}
