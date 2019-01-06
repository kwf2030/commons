package wechatbot

import (
  "errors"
  "os"
  "path"
  "strconv"
  "sync"
  "time"

  "github.com/kwf2030/commons/times"
)

const (
  StateCreated = iota

  // 已扫码未确认
  StateScanned

  // 已确认（正在登录）
  StateConfirmed

  // 登录成功（此时可以正常收发消息）
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

  // 退出（主动或被动）
  EventExit

  // 收到消息
  EventMsg
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

  // 未登录成功时会用时间戳作为key，保证bots中有记录且可查询这个Bot
  attrRandUin = "wechatbot.rand_uin"

  rootDir = "wechatbot"
  dumpDir = rootDir + "/dump/"

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
  // 所有Bot，int=>*Bot
  // 若处于Created/Scanned/Confirmed状态，key是时间戳，
  // 若处于Running和Stopped状态，key是uin，
  // 调用Bot.Release()会把Bot从bots中删除
  bots = &sync.Map{}

  dumpToFileEnabled = false
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

func eachBot(f func(*Bot) bool) {
  bots.Range(func(_, bot interface{}) bool {
    return f(bot.(*Bot))
  })
}

func EnableDumpToFile(enabled bool) {
  dumpToFileEnabled = enabled
}

func CountBots() int {
  i := 0
  eachBot(func(b *Bot) bool {
    i++
    return true
  })
  return i
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

type Bot struct {
  Attr *sync.Map

  Self     *Contact
  Contacts *Contacts

  StartTime time.Time
  StopTime  time.Time

  op  chan *op
  evt chan *Event

  req *req
}

func GetBotByUUID(uuid string) *Bot {
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

func GetBotByUin(uin int64) *Bot {
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

func Create() *Bot {
  ch := make(chan *op, 4)
  bot := &Bot{
    Attr: &sync.Map{},
    op:   ch,
    evt:  make(chan *Event, 8),
    req:  newReq(),
  }
  bot.req.bot = bot
  // 未获取到uin之前key是当前时间戳，
  // 无论登录成功还是失败，都会删除这个key，
  // 如果登录成功，会用uin存储这个Bot
  k := times.Timestamp()
  bot.Attr.Store(attrRandUin, k)
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
      bot.evt <- &Event{Type: EventSignInFailed, Err: e}
      close(bot.op)
      return
    }

    bots.Store(bot.req.Uin, bot)
    bot.StartTime = times.Now()
    bot.req.State = StateRunning
    bot.evt <- &Event{Type: EventSignInSuccess}
  }()
  return bot.evt
}

func (bot *Bot) Stop() {
  bot.StopTime = times.Now()
  bot.req.State = StateStopped
  bot.req.SignOut()
}

func (bot *Bot) Release() {
  bots.Delete(bot.req.Uin)
  bot.req.reset()
  bot.req.bot = nil
  bot.req.flow = nil
  bot.req.client = nil
  bot.req = nil
  bot.Attr = nil
  bot.Self = nil
  bot.Contacts = nil
  bot.op = nil
  bot.evt = nil
}

func (bot *Bot) updatePaths() {
  if bot.req.Uin == 0 {
    return
  }
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
    evt := &Event{Type: -1}
    switch op.what {
    case opAddMsg:
      evt.Type = EventMsg
      evt.Msg = op.msg
    case opModContact:
      bot.Contacts.Add(op.contact)
    case opDelContact:
      bot.Contacts.Remove(op.contact.UserName)
    case opQR:
      evt.Type = EventQRCode
      evt.Str = bot.req.QRCodeUrl
    case opSignIn:
      bot.updatePaths()
    case opInit:
      bot.Self = op.contact
    case opContacts:
      bot.Contacts = initContacts(op.contacts, bot)
    case opExit:
      evt.Type = EventExit
      bot.Stop()
    }
    if evt.Type != -1 {
      bot.evt <- evt
    }
  }
  // 如果syncCheck请求收到非零的响应，由它负责关闭op（谁发送谁关闭的原则），
  // 但evt是在这里是发送方，所以应该在此处关闭，
  // 不能放在Stop方法里（如果在Stop里面关闭了evt，就没法发送退出事件了）
  close(bot.evt)
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
  Int     int
  Str     string
  Contact *Contact
  Msg     *Message
}
