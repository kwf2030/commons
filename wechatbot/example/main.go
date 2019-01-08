package main

import (
  "bytes"
  "fmt"
  "os/exec"
  "runtime"
  "sync"
  "time"

  "github.com/kwf2030/commons/times"
  "github.com/kwf2030/commons/wechatbot"
)

type MsgHandler struct {
  // wechatbot.DefaultHandler实现了wechatbot.Handler接口，且具有默认行为，
  // 将DefaultHandler组合到struct中，关心哪个事件，实现接口的哪个方法即可，
  // 注意DefaultHandler是值，不是指针
  wechatbot.DefaultHandler
}

// 登录回调
func (h *MsgHandler) OnSignIn(ctx *wechatbot.HandlerContext, e error) {
  if e == nil {
    fmt.Println("登录成功")
  } else {
    fmt.Println("登录失败:", e.Error())
  }
}

// 退出回调
func (h *MsgHandler) OnSignOut(ctx *wechatbot.HandlerContext, _ error) {
  bot := ctx.Bot()
  var buf bytes.Buffer
  buf.WriteString("[%s] 已退出:\n")
  buf.WriteString("登录: %s\n")
  buf.WriteString("退出: %s\n")
  buf.WriteString("共在线 %.2f 小时\n")
  fmt.Printf(buf.String(), bot.Self.NickName,
    bot.StartTime.Format(times.DateTimeFormat),
    bot.StopTime.Format(times.DateTimeFormat),
    bot.StopTime.Sub(bot.StartTime).Hours())
}

// 二维码回调，需要扫码登录，
// qrcodeUrl是二维码的Url
func (h *MsgHandler) OnQRCode(ctx *wechatbot.HandlerContext, qrcodeUrl string) {
  // 方便起见，这里使用自带的方法下载二维码并自动打开，
  // 参数传空表示下载到系统临时目录，
  // 返回下载的完整路径
  p, _ := ctx.Bot().DownloadQRCode("")
  switch runtime.GOOS {
  case "windows":
    exec.Command("cmd.exe", "/c", p).Start()
  case "linux":
    exec.Command("eog", p).Start()
  default:
    fmt.Printf("二维码已保存至[%s]，请打开后扫码登录\n", p)
  }
}

// 消息回调
func (h *MsgHandler) OnMessage(ctx *wechatbot.HandlerContext, msg *wechatbot.Message, code int) {
  content := "<NULL>"
  if msg.Content != "" {
    content = msg.Content
  }

  // GroupMsgHandler解析后会给SpeakerUserName赋值，这时候才能区分群聊与单聊，
  // 如果没有添加GroupMsgHandler，就是原始消息，需要自己解析
  if msg.SpeakerUserName != "" {
    fmt.Printf("From:%s\nTo:%s\nSpeaker:%s\nType:%d\nContent:%s\n\n", msg.FromUserName, msg.ToUserName, msg.SpeakerUserName, msg.Type, content)
  } else {
    fmt.Printf("From:%s\nTo:%s\nType:%d\nContent:%s\n\n", msg.FromUserName, msg.ToUserName, msg.Type, content)
  }

  c := msg.GetFromContact()
  if c == nil || c.Type != wechatbot.ContactFriend {
    return
  }

  var reply string
  switch msg.Type {
  case wechatbot.MsgText:
    reply = "收到文本"

  case wechatbot.MsgImage:
    reply = "收到图片"

  case wechatbot.MsgAnimEmotion:
    reply = "收到动画表情"

  case wechatbot.MsgLink:
    reply = "收到链接"

  case wechatbot.MsgCard:
    reply = "收到名片"

  case wechatbot.MsgLocation:
    reply = "收到位置"

  case wechatbot.MsgVoice:
    reply = "收到语音"

  case wechatbot.MsgVideo:
    reply = "收到视频"
  }
  if reply == "" {
    return
  }
  msg.ReplyText(reply)
}

func main() {
  var wg sync.WaitGroup
  wg.Add(1)

  // 启用dump，将每个收到的数据作为一个文件写入wechatbot/dump目录内，
  // 对调试和分析数据非常有用
  wechatbot.EnableDumpToFile(true)

  // 创建Bot的时候依次传入Handler，
  // 回调函数不要阻塞，耗时操作放到goroutine中执行，
  // 添加验证消息（VerifyMsgHandler）和群消息（GroupMsgHandler）处理，注意顺序，
  // 如果想处理原始消息数据，可以不添加这两个Handler
  bot := wechatbot.Create(&wechatbot.VerifyMsgHandler{}, &wechatbot.GroupMsgHandler{}, &MsgHandler{})
  bot.Start()

  wg.Wait()
  time.Sleep(time.Second)
}
