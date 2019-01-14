package main

import (
  "bytes"
  "log"
  "os/exec"
  "runtime"
  "sync"
  "time"

  "github.com/kwf2030/commons/times"
  "github.com/kwf2030/commons/wxbot"
)

var wg sync.WaitGroup

type Handler struct {
  bot *wxbot.Bot
}

// 登录回调
func (h *Handler) OnSignIn(e error) {
  if e == nil {
    log.Println("sign in success")
  } else {
    log.Println("sign in failed:", e)
  }
}

// 退出回调
func (h *Handler) OnSignOut() {
  var buf bytes.Buffer
  buf.WriteString("[%s] is offline:\n")
  buf.WriteString("sign in: %s\n")
  buf.WriteString("sign out: %s\n")
  buf.WriteString("online for %.2f hours\n")
  log.Printf(buf.String(), h.bot.Self().NickName,
    h.bot.StartTime.Format(times.DateTimeFormat),
    h.bot.StopTime.Format(times.DateTimeFormat),
    h.bot.StopTime.Sub(h.bot.StartTime).Hours())
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
    log.Printf("qr code is saved to [%s], open it and scan for sign in\n", p)
  }
}

func (h *Handler) OnContact(c *wxbot.Contact, _ int) {
  log.Printf("OnContact: %s\n", c.NickName)
}

// 消息回调
func (h *Handler) OnMessage(msg *wxbot.Message, _ int) {
  if msg.Content == "" {
    msg.Content = "<NULL>"
  }
  to, from := "", ""
  c := msg.GetToContact()
  if c != nil {
    to = c.NickName
  }
  c = msg.GetFromContact()
  if c != nil {
    from = c.NickName
  }
  if msg.SpeakerUserName != "" {
    log.Printf("\nFrom: %s[%s](Group)\nTo: %s[%s]\nSpeaker: %s\nType: %d\nContent: %s\n", from, msg.FromUserName, to, msg.ToUserName, msg.SpeakerUserName, msg.Type, msg.Content)
  } else {
    log.Printf("\nFrom: %s[%s]\nTo: %s[%s]\nType: %d\nContent: %s\n", from, msg.FromUserName, to, msg.ToUserName, msg.Type, msg.Content)
  }
  if c == nil || c.Type != wxbot.ContactFriend {
    return
  }
  switch msg.Type {
  case wxbot.MsgText:
    msg.ReplyText("收到文本")
  case wxbot.MsgImage:
    msg.ReplyText("收到图片")
  case wxbot.MsgAnimEmotion:
    msg.ReplyText("收到动画表情")
  case wxbot.MsgLink:
    msg.ReplyText("收到链接")
  case wxbot.MsgCard:
    msg.ReplyText("收到名片")
  case wxbot.MsgLocation:
    msg.ReplyText("收到位置")
  case wxbot.MsgVoice:
    msg.ReplyText("收到语音")
  case wxbot.MsgVideo:
    msg.ReplyText("收到视频")
  }
}

func main() {
  wxbot.EnableDump(true)
  bot := wxbot.New()
  bot.Start(&Handler{bot: bot})
  wg.Add(1)
  wg.Wait()
  time.Sleep(time.Second)
}
