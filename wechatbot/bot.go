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

const opSignIn = 0x0001

var (
  ErrContactNotFound = errors.New("contact not found")

  ErrInvalidArgs = errors.New("invalid args")

  ErrInvalidState = errors.New("invalid state")

  ErrReq = errors.New("request failed")

  ErrResp = errors.New("response invalid")

  ErrTimeout = errors.New("timeout")

  ErrSignOut = errors.New("sign out")
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

  op       chan op
  pipeline *Pipeline

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

func Create(handlers ...Handler) *Bot {
  ch := make(chan op, 4)
  bot := &Bot{
    Attr: &sync.Map{},
    op:   ch,
    req:  newReq(),
  }
  bot.req.bot = bot
  // 未获取到uin之前key是当前时间戳，
  // 无论登录成功还是失败，都会删除这个key，
  // 如果登录成功，会用uin存储这个Bot
  k := times.Timestamp()
  bot.Attr.Store(attrRandUin, k)
  bot.pipeline = newPipeline(bot)
  bot.pipeline.AddLast(&DispatchHandler{})
  for _, h := range handlers {
    bot.pipeline.AddLast(h)
  }
  bots.Store(k, bot)
  return bot
}

// 返回的channel用来接收事件通知
func (bot *Bot) Start() {
  go bot.dispatch()
  go func() {
    bot.req.initFlow()
    _, e := bot.req.flow.Start(nil)

    // 不管登录成功还是失败，都要把临时的kv删除
    k, _ := bot.Attr.Load(attrRandUin)
    bots.Delete(k)

    if e != nil {
      // 登录Bot出现了问题或一直没扫描超时了
      bot.op <- op{what: opSignIn, data: e}
      close(bot.op)
      return
    }

    bots.Store(bot.req.Uin, bot)
    bot.StartTime = times.Now()
    bot.req.State = StateRunning
    bot.op <- op{what: opSignIn}
  }()
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
  bot.pipeline.Clear()
  bot.pipeline = nil
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
    switch op.what {
    case opSync:
      var data []byte
      if op.data != nil {
        data = op.data.([]byte)
      }
      ctx := bot.pipeline.head
      ctx.handler.OnData(ctx, data, op.syncCheck.code, op.syncCheck.selector)
    case opQR:
      ctx := bot.pipeline.head
      ctx.handler.OnQRCode(ctx, bot.req.QRCodeUrl)
    case opRedirect:
      bot.updatePaths()
    case opInit:
      bot.Self = op.data.(*Contact)
    case opContacts:
      bot.Contacts = initContacts(op.data.([]*Contact), bot)
    case opSignIn:
      var e error
      if op.data != nil {
        e = op.data.(error)
      }
      ctx := bot.pipeline.head
      ctx.handler.OnSignIn(ctx, e)
    case opSignOut:
      bot.Stop()
      var e error
      if op.syncCheck.code != 1101 {
        e = ErrSignOut
      }
      ctx := bot.pipeline.head
      ctx.handler.OnSignOut(ctx, e)
    }
  }
}

type op struct {
  what      int
  data      interface{}
  syncCheck syncCheckResp
}
