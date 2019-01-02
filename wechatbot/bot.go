package wechatbot

import (
  "errors"
  "math/rand"
  "net/http"
  "net/http/cookiejar"
  "net/url"
  "os"
  "path"
  "runtime"
  "strconv"
  "sync"
  "time"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
  "golang.org/x/net/publicsuffix"
)

const (
  StateCreated = iota

  // 已扫码未确认
  StateScanned

  // 已确认（正在登录）
  StateConfirmed

  // 登录成功（此时才可以正常收发消息）
  StateRunning

  // 停止/下线（手动、被动或异常）
  StateStopped

  // 超时
  StateTimeout
)

// Event Type
const (
  // 收到二维码
  EventQRCode = iota

  // 登录成功
  EventSignInSuccess

  // 登录失败
  EventSignInFailed

  // 主动退出
  EventSignOut

  // 被动退出
  EventOffline

  // 收到消息
  EventMsg

  // 添加好友（主动或被动）
  EventContactNew

  // 删除好友（主动或被动）
  EventContactDel

  // 好友资料更新
  EventContactMod

  // 进群（建群、主动或被动加群）
  EventGroupNew

  // 退群（主动或被动）
  EventGroupDel

  // 群资料更新（名称更新/群主变更/设置更新等）
  EventGroupMod
)

const (
  // 图片消息存放目录
  AttrImageDir = "wechatbot.image_dir"

  // 语音消息存放目录
  AttrVoiceDir = "wechatbot.voice_dir"

  // 视频消息存放目录
  AttrVideoDir = "wechatbot.video_dir"

  // 文件消息存放目录
  AttrFileDir = "wechatbot.file_dir"

  // 头像存放路径
  AttrAvatarPath = "wechatbot.avatar_path"

  // 持久化方案，如果禁用则Contact.Id永远不会有值，默认禁用，
  // 仅会对好友（使用备注实现）和群（使用群名称实现，改名会同步，同名没影响）持久化
  attrIdEnabled = "wechatbot.id_enabled"

  // 初始Id
  attrInitialId = "wechatbot.initial_id"

  // 未登录成功时会随机生成uin作为key，保证bots中有记录且可查询这个Bot
  attrRandUin = "wechatbot.rand_uin"

  rootDir     = "wechatbot"
  contentType = "application/json; charset=UTF-8"
  userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
)

var (
  ErrContactNotFound = errors.New("contact not found")

  ErrInvalidArgs = errors.New("invalid args")

  ErrInvalidState = errors.New("invalid state")

  ErrReq = errors.New("request failed")

  ErrResp = errors.New("response invalid")

  ErrTimeout = errors.New("timeout")
)

// 所有Bot，int=>*Bot
// 若处于Created/Scanned/Confirmed状态，key是随机生成的，
// 若处于Running和Stopped状态，key是uin，
// 调用Bot.Release()会把Bot从bots中删除
var bots = &sync.Map{}

func init() {
  e := os.MkdirAll(rootDir, os.ModePerm)
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

func eachBot(f func(*Bot) bool) {
  bots.Range(func(_, bot interface{}) bool {
    if b, ok := bot.(*Bot); ok {
      return f(b)
    }
    return true
  })
}

func RunningBots() []*Bot {
  ret := make([]*Bot, 0, 2)
  eachBot(func(b *Bot) bool {
    if b.req.State == StateRunning {
      ret = append(ret, b)
    }
    return true
  })
  return ret
}

func CountBots() int {
  i := 0
  eachBot(func(b *Bot) bool {
    i++
    return true
  })
  return i
}

func FindBotByUUID(uuid string) *Bot {
  if uuid == "" {
    return nil
  }
  var ret *Bot
  eachBot(func(b *Bot) bool {
    if b.req != nil && b.req.UUID == uuid {
      ret = b
      return false
    }
    return true
  })
  return ret
}

func FindBotByUin(uin int64) *Bot {
  var ret *Bot
  eachBot(func(b *Bot) bool {
    if b.req != nil && b.req.Uin == uin {
      ret = b
      return false
    }
    return true
  })
  return ret
}

type Bot struct {
  Attr *sync.Map

  Self     *Contact
  Contacts *Contacts

  StartTime time.Time
  StopTime  time.Time

  evt chan *Event
  op  chan *op

  req *req
}

func CreateBot(enableId bool) *Bot {
  ch := make(chan *op, runtime.NumCPU()+1)
  bot := &Bot{
    Attr: &sync.Map{},
    evt:  make(chan *Event, cap(ch)),
    op:   ch,
    req:  newReq(),
  }
  bot.req.bot = bot
  // 未获取到uin之前key是随机的，
  // 无论登录成功还是失败，都会删除这个key，
  // 如果登录成功，会用uin存储这个Bot
  k := rand.Int()
  bot.Attr.Store(attrRandUin, k)
  bot.Attr.Store(attrIdEnabled, enableId)
  bots.Store(k, bot)
  return bot
}

// 返回的channel用来接收事件通知
func (bot *Bot) Start() <-chan *Event {
  go bot.dispatch()
  go func() {
    bot.req.initFlow()
    _, e := bot.req.flow.Start(nil)

    // 不管登录成功还是失败，都要把临时的kv删除
    k, _ := bot.Attr.Load(attrRandUin)
    bots.Delete(k)

    if e != nil {
      // 登录Bot出现了问题或一直没扫描超时了
      close(bot.op)
      bot.Release()
      bot.evt <- &Event{Type: EventSignInFailed, Err: e}
      return
    }

    bot.StartTime = times.Now()
    bot.req.State = StateRunning
    bots.Store(bot.req.Uin, bot)
    bot.evt <- &Event{Type: EventSignInSuccess}
  }()
  return bot.evt
}

func (bot *Bot) GetAttrString(attr string, defaultValue string) string {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.String(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrInt(attr string, defaultValue int) int {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Int(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrInt64(attr string, defaultValue int64) int64 {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Int64(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrUint(attr string, defaultValue uint) uint {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Uint(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrUint64(attr string, defaultValue uint64) uint64 {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Uint64(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrBool(attr string, defaultValue bool) bool {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Bool(v)
  }
  return defaultValue
}

func (bot *Bot) GetAttrBytes(attr string) []byte {
  if v, ok := bot.Attr.Load(attr); ok {
    switch ret := v.(type) {
    case []byte:
      return ret
    case string:
      return []byte(ret)
    }
  }
  return nil
}

func (bot *Bot) updatePaths() {
  uin := strconv.FormatInt(bot.req.Uin, 10)
  dir := path.Join(rootDir, uin, times.NowStrf(times.DateFormat))
  e := os.MkdirAll(dir, os.ModePerm)
  if e != nil {
    return
  }

  image := path.Join(dir, "image")
  e = os.MkdirAll(image, os.ModePerm)
  if e == nil {
    bot.Attr.Store(AttrImageDir, image)
  }

  voice := path.Join(dir, "voice")
  e = os.MkdirAll(voice, os.ModePerm)
  if e == nil {
    bot.Attr.Store(AttrVoiceDir, voice)
  }

  video := path.Join(dir, "video")
  e = os.MkdirAll(video, os.ModePerm)
  if e == nil {
    bot.Attr.Store(AttrVideoDir, video)
  }

  file := path.Join(dir, "file")
  e = os.MkdirAll(file, os.ModePerm)
  if e == nil {
    bot.Attr.Store(AttrFileDir, file)
  }

  bot.Attr.Store(AttrAvatarPath, path.Join(rootDir, uin, "avatar.jpg"))
}

func (bot *Bot) dispatch() {
  for op := range bot.op {
    // evt := &Event{}

    switch op.what {
    case 1:
    case 2:
    case 3:
    case 4:
    case 5:
    case 6:
    case 7:
    case 8:
    case 9:
    }
  }

  // op不用关闭，在Logout之后，syncCheck请求会收到非零的响应，
  // 由它负责关闭op（谁发送谁关闭的原则），
  // 但evt是在dispatch的时候发送给调用方的，所以应该在此处关闭，
  // 不能放在Stop方法里（如果在Stop里面关闭了evt，那就没法发送TerminateOp事件了），
  // 而且会引起"send on closed channel"的panic
  close(bot.evt)
}

/*func (bot *Bot) dispatch() {
  for o := range bot.op {
    op := &Op{What: o.What}
    switch o.What {
    case MsgOp:
      op.Msg = mapToMessage(o.Data.(map[string]interface{}), bot)

    case ContactModOp:
      op.Contact = bot.modContact(o.Data.(map[string]interface{}))

    case ContactDelOp:
      c := mapToContact(o.Data.(map[string]interface{}), bot)
      bot.Contacts.Remove(c.UserName)
      op.Contact = c

    case ContactSelfOp:
      bot.updatePaths()
      bot.Self = mapToContact(o.Data.(map[string]interface{}), bot)

    case ContactListOp:
      bot.Contacts = initContacts(mapsToContacts(o.Data.([]map[string]interface{}), bot), bot)

    case TerminateOp:
      bot.Stop()
    }
    bot.opForward <- op
  }
}*/

/*func (bot *Bot) modContact(m map[string]interface{}) *Contact {
  c := mapToContact(m, bot)
  if !bot.Attr[AttrPersistentIDEnabled].(bool) {
    bot.Contacts.Add(c)
    return c
  }
  switch c.Flag {
  case ContactFriend:
    // 如果ID是空，说明是新联系人
    if c.ID == "" {
      // 关闭好友验证的情况下，被添加好友时会收到此类消息，
      // ContactModOp会先于MsgOp事件发出，所以收到MsgOp时，该联系人一定已存在
      c.ID = strconv.FormatUint(bot.Contacts.nextID(), 10)
      c.CreateTime = times.Now()
      bot.req.Remark(c.UserName, c.ID)
    }
  case ContactGroup:

  case ContactSystem:
    n, ok := internalIDs[c.UserName]
    if !ok {
      n = uint64(len(internalIDs) + 1)
      internalIDs[c.UserName] = n
    }
    c.ID = strconv.FormatUint(bot.Contacts.initialID()+n, 10)
  }
  bot.Contacts.Add(c)
  return c
}*/

func (bot *Bot) Stop() {
  bot.StopTime = times.Now()
  bot.req.State = StateStopped
  bot.req.SignOut()
}

func (bot *Bot) Release() {
  bots.Delete(bot.req.Uin)
  bot.req.reset()
  bot.req.flow = nil
  bot.req.client = nil
  bot.req = nil
  bot.Attr = nil
  bot.Self = nil
  bot.Contacts = nil
  bot.op = nil
}

type req struct {
  bot    *Bot
  flow   *flow.Flow
  client *http.Client
  *session
}

func newReq() *req {
  jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
  sess := &session{}
  sess.reset()
  return &req{
    session: sess,
    flow:    flow.NewFlow(0),
    client: &http.Client{
      Jar:     jar,
      Timeout: time.Minute * 2,
    },
  }
}

func (r *req) initFlow() {
  uuid := &uuidReq{r}
  scanState := &scanStateReq{r}
  login := &loginReq{r}
  init := &initReq{r}
  statusNotify := &statusNotifyReq{r}
  contactList := &contactListReq{r}
  syn := &syncReq{r}
  r.flow.AddLast(uuid, "uuid")
  r.flow.AddLast(scanState, "scan_state")
  r.flow.AddLast(login, "login")
  r.flow.AddLast(init, "init")
  r.flow.AddLast(statusNotify, "status_notify")
  r.flow.AddLast(contactList, "contact_list")
  r.flow.AddLast(syn, "sync")
}

func (r *req) cookie(key string) string {
  if key == "" {
    return ""
  }
  addr, _ := url.Parse(r.BaseUrl)
  arr := r.client.Jar.Cookies(addr)
  for _, c := range arr {
    if c.Name == key {
      return c.Value
    }
  }
  return ""
}

type session struct {
  SyncCheckHost string
  Host          string
  Referer       string
  BaseUrl       string
  State         int
  UUID          string
  QRCodeUrl     string
  RedirectUrl   string
  Uin           int64
  Sid           string
  SKey          string
  PassTicket    string
  BaseReq       *baseReq
  UserName      string
  AvatarUrl     string
  SyncKeys      *syncKeys
  WuFile        int
}

func (s *session) reset() {
  s.SyncCheckHost = "webpush.weixin.qq.com"
  s.Host = "wx.qq.com"
  s.Referer = "https://wx.qq.com/"
  s.BaseUrl = "https://wx.qq.com/cgi-bin/mmwebwx-bin"
  s.State = StateCreated
  s.UUID = ""
  s.QRCodeUrl = ""
  s.RedirectUrl = ""
  s.BaseReq = nil
  s.Uin = 0
  s.Sid = ""
  s.SKey = ""
  s.PassTicket = ""
  s.UserName = ""
  s.AvatarUrl = ""
  s.SyncKeys = nil
  s.WuFile = 0
}

type op struct {
  what     int
  contact  *Contact
  contacts []*Contact
  msg      *Message

  syncCheckCode     int
  syncCheckSelector int
}

type Event struct {
  Err     error
  Type    int
  SubType int
  Int1    int
  Int2    int
  Str1    string
  Str2    string
  Contact *Contact
  Msg     *Message
}

func deviceId() string {
  return "e" + timestampStringL(15)
}

func timestampString13() string {
  return timestampStringL(13)
}

func timestampString10() string {
  return timestampStringL(10)
}

func timestampStringL(l int) string {
  s := strconv.FormatInt(times.Timestamp(), 10)
  if len(s) > l {
    return s[:l]
  }
  return s
}

func timestampStringR(l int) string {
  s := strconv.FormatInt(times.Timestamp(), 10)
  i := len(s) - l
  if i > 0 {
    return s[i:]
  }
  return s
}
