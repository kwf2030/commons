package main

import (
  "bytes"
  "fmt"
  "os/exec"
  "runtime"
  "sync"
  "time"

  "github.com/kwf2030/commons/times"
  "github.com/kwf2030/commons/wxbot"
)

var wg sync.WaitGroup

type Handler struct {
  // wxbot.DefaultHandler实现了wxbot.Handler接口的所有方法（空实现），
  // 组合进来就不用再实现不需要的方法了
  wxbot.DefaultHandler

  bot *wxbot.Bot
}

// 登录回调
func (h *Handler) OnSignIn(e error) {
  if e == nil {
    fmt.Println("登录成功")
  } else {
    fmt.Println("登录失败:", e)
  }
}

// 退出回调
func (h *Handler) OnSignOut(error) {
  var buf bytes.Buffer
  buf.WriteString("[%s] 已退出:\n")
  buf.WriteString("登录: %s\n")
  buf.WriteString("退出: %s\n")
  buf.WriteString("共在线 %.2f 小时\n")
  fmt.Printf(buf.String(), h.bot.Self().NickName,
    h.bot.StartTime().Format(times.DateTimeFormat),
    h.bot.StopTime().Format(times.DateTimeFormat),
    h.bot.StopTime().Sub(h.bot.StartTime()).Hours())
  wg.Done()
}

// 二维码回调，需要扫码登录，
// qrCodeUrl是二维码的链接
func (h *Handler) OnQRCode(qrcodeUrl string) {
  p, _ := h.bot.DownloadQRCode("")
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
func (h *Handler) OnMessage(msg *wxbot.Message, code int) {
  content := "<NULL>"
  if msg.Content != "" {
    content = msg.Content
  }

  if msg.SpeakerUserName != "" {
    fmt.Printf("From:%s\nTo:%s\nSpeaker:%s\nType:%d\nContent:%s\n\n", msg.FromUserName, msg.ToUserName, msg.SpeakerUserName, msg.Type, content)
  } else {
    fmt.Printf("From:%s\nTo:%s\nType:%d\nContent:%s\n\n", msg.FromUserName, msg.ToUserName, msg.Type, content)
  }

  c := msg.GetFromContact()
  if c == nil || c.Type != wxbot.ContactFriend {
    return
  }

  var reply string
  switch msg.Type {
  case wxbot.MsgText:
    reply = "收到文本"

  case wxbot.MsgImage:
    reply = "收到图片"

  case wxbot.MsgAnimEmotion:
    reply = "收到动画表情"

  case wxbot.MsgLink:
    reply = "收到链接"

  case wxbot.MsgCard:
    reply = "收到名片"

  case wxbot.MsgLocation:
    reply = "收到位置"

  case wxbot.MsgVoice:
    reply = "收到语音"

  case wxbot.MsgVideo:
    reply = "收到视频"
  }
  if reply == "" {
    return
  }
  msg.ReplyText(reply)
}

func main() {
  bot := wxbot.New()
  bot.Start(&Handler{bot: bot})
  wg.Add(1)
  wg.Wait()
  time.Sleep(time.Second)
}
