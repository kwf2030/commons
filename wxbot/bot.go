package wxbot

import (
  "errors"
  "fmt"
  "net/http"
  "net/http/cookiejar"
  "os"
  "path"
  "strconv"
  "strings"
  "sync"
  "time"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/times"
  "golang.org/x/net/publicsuffix"
)

const (
  stateUnknown = iota

  // 已扫码未确认
  StateScan

  // 等待确认超时
  StateScanTimeout

  // 已确认（正在登录）
  StateConfirm

  // 登录成功（此时可以正常收发消息）
  StateRunning

  // 已下线（主动、被动或异常）
  StateStop
)

const (
  // 图片消息存放目录
  attrImageDir = "wechatbot.image_dir"

  // 语音消息存放目录
  attrVoiceDir = "wechatbot.voice_dir"

  // 视频消息存放目录
  attrVideoDir = "wechatbot.video_dir"

  // 文件消息存放目录
  attrFileDir = "wechatbot.file_dir"

  // 头像存放路径
  attrAvatarPath = "wechatbot.avatar_path"

  // 正在登录时用时间戳作为key，保证bots中有记录且可查询这个Bot
  attrRandUin = "wechatbot.rand_uin"

  rootDir = "wechatbot"
  dumpDir = rootDir + "/dump/"

  contentType = "application/json; charset=UTF-8"
  userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
)

var (
  ErrInvalidArgs = errors.New("invalid args")

  ErrInvalidState = errors.New("invalid state")

  ErrReq = errors.New("request failed")

  ErrResp = errors.New("response invalid")

  ErrScanTimeout = errors.New("scan timeout")

  ErrSignOut = errors.New("sign out")

  ErrContactNotFound = errors.New("contact not found")
)

var (
  dumpEnabled = false

  botsMutex = &sync.RWMutex{}
  bots      = make(map[int64]*Bot, 4)
)

func init() {
  e := os.MkdirAll(dumpDir, os.ModePerm)
  if e != nil {
    return
  }
  updatePaths()
}

func updatePaths() {
  time.AfterFunc(times.UntilTomorrow(), func() {
    for _, b := range RunningBots() {
      b.updatePaths()
    }
    updatePaths()
  })
}

func EachBot(f func(*Bot) bool) {
  arr := make([]*Bot, 0, 2)
  botsMutex.RLock()
  for _, v := range bots {
    arr = append(arr, v)
  }
  botsMutex.RUnlock()
  for _, v := range arr {
    if !f(v) {
      break
    }
  }
}

func CountBots() int {
  l := 0
  botsMutex.RLock()
  l = len(bots)
  botsMutex.RUnlock()
  return l
}

func RunningBots() []*Bot {
  ret := make([]*Bot, 0, 4)
  botsMutex.RLock()
  for _, v := range bots {
    if v.session.State == StateRunning {
      ret = append(ret, v)
    }
  }
  botsMutex.RUnlock()
  if len(ret) == 0 {
    return nil
  }
  return ret
}

func EnableDump(enabled bool) {
  dumpEnabled = enabled
}

type Bot struct {
  callback Handler

  client  *http.Client
  session *session
  req     *wxReq

  signInPipeline *pipeline
  syncPipeline   *pipeline

  self     *Contact
  contacts *Contacts

  attr *sync.Map

  startTime time.Time
  stopTime  time.Time
}

func New() *Bot {
  jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
  s := &session{}
  s.init()
  bot := &Bot{
    client: &http.Client{
      Jar:     jar,
      Timeout: time.Minute * 2,
    },
    session:        s,
    signInPipeline: newPipeline(),
    syncPipeline:   newPipeline(),
    attr:           &sync.Map{},
  }
  bot.req = &wxReq{bot}
  // 未获取到uin之前key是当前时间戳，
  // 无论登录成功还是失败，都会删除这个key，
  // 如果登录成功，会用uin存储这个Bot
  k := times.Timestamp()
  bot.attr.Store(attrRandUin, k)
  botsMutex.Lock()
  bots[k] = bot
  botsMutex.Unlock()
  return bot
}

func GetBotByUUID(uuid string) *Bot {
  if uuid == "" {
    return nil
  }
  var ret *Bot
  EachBot(func(b *Bot) bool {
    if b.session.UUID == uuid {
      ret = b
      return false
    }
    return true
  })
  return ret
}

func GetBotByUin(uin int64) *Bot {
  var ret *Bot
  EachBot(func(b *Bot) bool {
    if b.session.Uin == uin {
      ret = b
      return false
    }
    return true
  })
  return ret
}

func (bot *Bot) StartTime() time.Time {
  return bot.startTime
}

func (bot *Bot) StopTime() time.Time {
  return bot.stopTime
}

func (bot *Bot) Self() *Contact {
  return bot.self
}

func (bot *Bot) Contacts() *Contacts {
  return bot.contacts
}

func (bot *Bot) Start(handler Handler) {
  if handler == nil {
    handler = &DefaultHandler{}
  }
  bot.callback = handler
  bot.syncPipeline.AddLast("dispatch", &dispatchHandler{bot}).
    AddLast("verify", &verifyMsgHandler{bot}).
    AddLast("group", &groupMsgHandler{bot})
  bot.signInPipeline.AddLast("qr", &qrReq{bot}).
    AddLast("scan", &scanReq{bot}).
    AddLast("redirect", &redirectReq{bot}).
    AddLast("init", &initReq{bot}).
    AddLast("notify", &notifyReq{bot}).
    AddLast("contacts", &contactsReq{bot}).
    AddLast("sync", &syncReq{bot})
  bot.signInPipeline.Fire(event{})
  if k, ok := bot.attr.Load(attrRandUin); ok {
    botsMutex.Lock()
    delete(bots, k.(int64))
    botsMutex.Unlock()
  }
}

func (bot *Bot) Stop() {
  bot.stopTime = times.Now()
  bot.session.State = StateStop
  bot.req.SignOut()
}

func (bot *Bot) updatePaths() {
  if bot.session.Uin == 0 {
    return
  }
  uin := strconv.FormatInt(bot.session.Uin, 10)
  dir := path.Join(rootDir, uin, times.NowStrf(times.DateFormat))
  e := os.MkdirAll(dir, os.ModePerm)
  if e != nil {
    return
  }

  image := path.Join(dir, "image")
  e = os.MkdirAll(image, os.ModePerm)
  if e == nil {
    bot.attr.Store(attrImageDir, image)
  }

  voice := path.Join(dir, "voice")
  e = os.MkdirAll(voice, os.ModePerm)
  if e == nil {
    bot.attr.Store(attrVoiceDir, voice)
  }

  video := path.Join(dir, "video")
  e = os.MkdirAll(video, os.ModePerm)
  if e == nil {
    bot.attr.Store(attrVideoDir, video)
  }

  file := path.Join(dir, "file")
  e = os.MkdirAll(file, os.ModePerm)
  if e == nil {
    bot.attr.Store(attrFileDir, file)
  }

  bot.attr.Store(attrAvatarPath, path.Join(rootDir, uin, "avatar.jpg"))
}

type session struct {
  Host          string
  SyncCheckHost string
  Referer       string
  BaseUrl       string

  State int

  UUID      string
  QRCodeUrl string

  RedirectUrl string

  SKey       string
  Sid        string
  Uin        int64
  PassTicket string
  BaseReq    baseReq

  SyncKey   syncKey
  UserName  string
  AvatarUrl string

  WuFile int
}

func (s *session) init() {
  s.Host = "wx.qq.com"
  s.SyncCheckHost = "webpush.weixin.qq.com"
  s.Referer = "https://wx.qq.com/"
  s.BaseUrl = "https://wx.qq.com/cgi-bin/mmwebwx-bin"
}

type baseReq struct {
  DeviceId string `json:"DeviceID"`
  Sid      string `json:"Sid"`
  SKey     string `json:"Skey"`
  Uin      int64  `json:"Uin"`
}

type syncKeyItem struct {
  Key int
  Val int
}

type syncKey struct {
  Count int
  List  []syncKeyItem
}

func parseSyncKey(data []byte) syncKey {
  count, _ := jsonparser.GetInt(data, "Count")
  if count <= 0 {
    return syncKey{}
  }
  arr := make([]syncKeyItem, 0, count)
  jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, i int, e error) {
    if e != nil {
      return
    }
    key, _ := jsonparser.GetInt(v, "Key")
    val, _ := jsonparser.GetInt(v, "Val")
    arr = append(arr, syncKeyItem{int(key), int(val)})
  }, "List")
  if len(arr) == 0 {
    return syncKey{}
  }
  return syncKey{Count: int(count), List: arr}
}

func (sk *syncKey) expand() string {
  var sb strings.Builder
  n := sk.Count - 1
  for i := 0; i <= n; i++ {
    item := sk.List[i]
    fmt.Fprintf(&sb, "%d_%d", item.Key, item.Val)
    if i != n {
      sb.WriteString("|")
    }
  }
  return sb.String()
}
