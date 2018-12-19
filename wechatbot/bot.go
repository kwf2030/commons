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
  Unknown = iota

  Created

  // 已扫码未确认
  Scanned

  // 已确认（正在登录）
  Confirmed

  // 登录成功（此时才可以正常收发消息）
  Running

  // 停止/下线（手动、被动或异常）
  Stopped

  // 超时
  Timeout
)

const (
  // 图片消息存放目录
  AttrDirImage = "wechatbot.attr.dir_image"

  // 语音消息存放目录
  AttrDirVoice = "wechatbot.attr.dir_voice"

  // 视频消息存放目录
  AttrDirVideo = "wechatbot.attr.dir_video"

  // 文件消息存放目录
  AttrDirFile = "wechatbot.attr.dir_file"

  // 头像存放路径
  AttrPathAvatar = "wechatbot.attr.path_avatar"

  // 持久化ID方案，如果禁用则Contact.ID永远不会有值，默认禁用，
  // 联系人持久化使用备注实现，群持久化使用群名称实现（改名会同步，同名没影响），公众号不会持久化
  AttrPersistentIDEnabled = "wechatbot.attr.persistent_id_enabled"

  // 未登录成功时会随机生成key，保证bots中有记录且可查询这个Bot
  attrBotPlaceHolder = "wechatbot.attr.bot_place_holder"

  // 起始ID
  attrInitialID = "wechatbot.attr.initial_id"
)

const (
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

var (
  once = sync.Once{}

  // 所有Bot，int=>*Bot
  // 若处于Created/Scanned/Confirmed状态，key是随机生成的，
  // 若处于Running和Stopped状态，key是uin，
  // 调用Bot.Release()会把Bot从bots中删除
  bots = sync.Map{}
)

func updatePaths() {
  now := times.Now()
  next := now.Add(time.Hour * 24)
  next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
  time.AfterFunc(next.Sub(now), func() {
    for _, b := range RunningBots() {
      b.updatePaths()
    }
    go updatePaths()
  })
}

func eachBot(f func(b *Bot) bool) {
  bots.Range(func(_, v interface{}) bool {
    if vv, ok := v.(*Bot); ok {
      return f(vv)
    }
    return true
  })
}

func RunningBots() []*Bot {
  ret := make([]*Bot, 0, 2)
  eachBot(func(b *Bot) bool {
    if b.State == Running {
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
    if b.req != nil && b.req.uuid == uuid {
      ret = b
      return false
    }
    return true
  })
  return ret
}

func FindBotByUin(uin int) *Bot {
  var ret *Bot
  eachBot(func(b *Bot) bool {
    if b.req != nil && b.req.uin == uin {
      ret = b
      return false
    }
    return true
  })
  return ret
}

type Bot struct {
  Attr map[string]interface{}

  Self     *Contact
  Contacts *Contacts

  StartTime    time.Time
  StartTimeStr string
  StopTime     time.Time
  StopTimeStr  string

  State int

  opForward chan *Op
  op        chan *op

  req *req
}

func CreateBot(enablePersistentID bool) *Bot {
  once.Do(func() {
    e := os.Mkdir("data", os.ModePerm)
    if e != nil {
      return
    }
    updatePaths()
  })
  ch := make(chan *op, runtime.NumCPU()+1)
  bot := &Bot{
    Attr:      make(map[string]interface{}),
    State:     Created,
    opForward: make(chan *Op, cap(ch)),
    op:        ch,
  }
  bot.req = newReq(bot, ch)
  // 未获取到uin之前key是随机的，
  // 登录失败或成功之后会删除这个key
  k := rand.Int()
  bot.Attr[attrBotPlaceHolder] = k
  bot.Attr[AttrPersistentIDEnabled] = enablePersistentID
  bots.Store(k, bot)
  return bot
}

// qrChan为接收二维码URL的channel，
// 返回的channel用来接收事件和消息通知，
// Start方法会一直阻塞到登录成功可以开始收发消息为止
func (bot *Bot) Start(qrChan chan<- string) (<-chan *Op, error) {
  if qrChan == nil {
    return nil, ErrInvalidArgs
  }

  // 监听事件和消息
  go bot.dispatch()

  bot.req.initFlow()
  _, e := bot.req.flow.Start(qrChan)

  // 不管登录成功还是失败，都要把临时的kv删除
  bots.Delete(bot.Attr[attrBotPlaceHolder])

  if e != nil {
    // 登录Bot出现了问题或一直没扫描超时了
    close(bot.op)
    bot.Release()
    return nil, e
  }

  t := times.Now()
  bot.StartTime = t
  bot.StartTimeStr = t.Format(times.DateTimeSFormat)
  bot.updatePaths()
  bot.State = Running
  bots.Store(bot.Self.Uin, bot)
  return bot.opForward, nil
}

func (bot *Bot) GetAttrString(attr string) string {
  return conv.String(bot.Attr, attr)
}

func (bot *Bot) GetAttrInt(attr string) int {
  return conv.Int(bot.Attr, attr)
}

func (bot *Bot) GetAttrUint64(attr string) uint64 {
  return conv.Uint64(bot.Attr, attr)
}

func (bot *Bot) GetAttrBool(attr string) bool {
  return conv.Bool(bot.Attr, attr)
}

func (bot *Bot) GetAttrBytes(attr string) []byte {
  if v, ok := bot.Attr[attr]; ok {
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
  dir := path.Join("data", strconv.Itoa(bot.Self.Uin), times.NowStrf(times.DateFormat))
  e := os.MkdirAll(dir, os.ModePerm)
  if e != nil {
    return
  }

  di := path.Join(dir, "image")
  dvo := path.Join(dir, "voice")
  dvi := path.Join(dir, "video")
  df := path.Join(dir, "file")

  os.MkdirAll(di, os.ModePerm)
  os.MkdirAll(dvo, os.ModePerm)
  os.MkdirAll(dvi, os.ModePerm)
  os.MkdirAll(df, os.ModePerm)

  bot.Attr[AttrPathAvatar] = path.Join(path.Dir(dir), "avatar.jpg")
  bot.Attr[AttrDirImage] = di
  bot.Attr[AttrDirVoice] = dvo
  bot.Attr[AttrDirVideo] = dvi
  bot.Attr[AttrDirFile] = df
}

func (bot *Bot) dispatch() {
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
      bot.Self = mapToContact(o.Data.(map[string]interface{}), bot)

    case ContactListOp:
      bot.Contacts = initContacts(mapsToContacts(o.Data.([]map[string]interface{}), bot), bot)

    case TerminateOp:
      bot.Stop()
    }
    // 事件转发
    bot.opForward <- op
  }

  // op不用关闭，在Logout之后，syncCheck请求会收到非零的响应，
  // 由它负责关闭op（谁发送谁关闭的原则），
  // 但opForward是在dispatch的时候发送给调用方的，所以应该在此处关闭，
  // 不能放在Stop方法里，因为如果在Stop里面关闭了op，那就没法发送TerminateOp事件了，
  // 而且会引起"send on closed channel"的panic
  close(bot.opForward)
}

func (bot *Bot) modContact(m map[string]interface{}) *Contact {
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
}

func (bot *Bot) Stop() {
  bot.State = Stopped
  t := times.Now()
  bot.StopTime = t
  bot.StopTimeStr = t.Format(times.DateTimeSFormat)
  bot.req.Logout()
}

func (bot *Bot) Release() {
  bots.Delete(bot.req.uin)
  bot.req.reset()
  bot.req.flow = nil
  bot.req.client = nil
  bot.req = nil
  bot.Attr = nil
  bot.Self = nil
  bot.Contacts = nil
  bot.State = Unknown
  bot.op = nil
}

type op struct {
  What int
  Data interface{}
}

type Op struct {
  What    int
  Msg     *Message
  Contact *Contact
}

type req struct {
  bot    *Bot
  op     chan<- *op
  flow   *flow.Flow
  client *http.Client
  *session
}

func newReq(bot *Bot, op chan<- *op) *req {
  jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
  s := &session{}
  s.reset()
  return &req{
    session: s,
    bot:     bot,
    op:      op,
    flow:    flow.NewFlow(0),
    client: &http.Client{
      Jar:     jar,
      Timeout: time.Minute * 2,
    },
  }
}

func (r *req) initFlow() {
  uuid := &UUIDReq{r}
  scanState := &ScanStateReq{r}
  login := &LoginReq{r}
  init := &InitReq{r}
  statusNotify := &StatusNotifyReq{r}
  contactList := &ContactListReq{r}
  syn := &SyncReq{r}
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
  addr, _ := url.Parse(r.baseURL)
  arr := r.client.Jar.Cookies(addr)
  for _, c := range arr {
    if c.Name == key {
      return c.Value
    }
  }
  return ""
}

type session struct {
  referer       string
  host          string
  syncCheckHost string
  baseURL       string

  uuid        string
  redirectURL string
  uin         int
  sid         string
  skey        string
  passTicket  string
  payload     map[string]interface{}

  userName  string
  avatarURL string
  syncKey   map[string]interface{}

  wuFile int
}

func (s *session) reset() {
  s.referer = "https://wx.qq.com/"
  s.host = "wx.qq.com"
  s.syncCheckHost = "webpush.weixin.qq.com"
  s.baseURL = "https://wx.qq.com/cgi-bin/mmwebwx-bin"
  s.uuid = ""
  s.redirectURL = ""
  s.uin = 0
  s.sid = ""
  s.skey = ""
  s.passTicket = ""
  s.payload = nil
  s.userName = ""
  s.avatarURL = ""
  s.syncKey = nil
  s.wuFile = 0
}

func timestamp() int64 {
  return times.Now().UnixNano()
}

func timestampStringN(l int, prefix, suffix string) string {
  s := strconv.FormatInt(timestamp(), 10)
  if len(s) <= l {
    return prefix + s + suffix
  }
  return prefix + s[:l] + suffix
}

func timestampString10() string {
  return timestampStringN(10, "", "")
}

func timestampString13() string {
  return timestampStringN(13, "", "")
}

func deviceID() string {
  return timestampStringN(15, "e", "")
}

func randStringN(l int) string {
  s := strconv.FormatInt(timestamp(), 10)
  if len(s) <= l {
    return s
  }
  return s[len(s)-l:]
}
