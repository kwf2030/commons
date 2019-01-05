package main

import (
  "bytes"
  "fmt"
  "os/exec"
  "runtime"
  "time"

  "github.com/kwf2030/commons/times"
  "github.com/kwf2030/commons/wechatbot"
)

func main() {
  // 启用dump，将每个收到的数据作为一个文件写入wechatbot/dump目录内，
  // 调试分析数据时非常有用
  wechatbot.EnableDumpToFile(true)

  bot := wechatbot.CreateBot(false)
  event := bot.Start()

  // 不要阻塞消息接收的channel，
  // 耗时操作放到goroutine中执行
  for evt := range event {
    switch evt.Type {
    // 收到消息
    case wechatbot.EventMsg:
      processMsg(evt.Msg)

    // 被添加好友，
    // 自动通过验证、添加到联系人并备注
    case wechatbot.EventFriendNew:
      bot.Accept(evt.Contact)
      evt.Msg.ReplyText("你好，朋友")

    // 获取到二维码，需扫码登录
    // evt.Str为二维码链接
    case wechatbot.EventQRCode:
      // 这里使用自带的方法下载二维码并自动打开，
      // 参数为完整路径，如果为空就下载到系统临时目录，
      // 返回二维码图片的完整路径
      p, _ := bot.DownloadQRCode("")
      switch runtime.GOOS {
      case "windows":
        exec.Command("cmd.exe", "/c", p).Start()
      case "linux":
        exec.Command("eog", p).Start()
      default:
        fmt.Printf("二维码已保存至[%s]，请打开后扫码登录", p)
      }

    case wechatbot.EventSignInSuccess:
      fmt.Println("登录成功")

    case wechatbot.EventSignInFailed:
      fmt.Println("登录失败")
    }
  }

  var buf bytes.Buffer
  buf.WriteString("WeChatBot[%s] run stat:\n")
  buf.WriteString("  started at: %s\n")
  buf.WriteString("  stopped at: %s\n")
  buf.WriteString("  totally online for %.2f hours\n")
  fmt.Printf(buf.String(), bot.Self.NickName,
    bot.StartTime.Format(times.DateTimeFormat),
    bot.StopTime.Format(times.DateTimeFormat),
    bot.StopTime.Sub(bot.StartTime).Hours())

  // 退出前等待1秒左右，以便退出请求和各种清理操作完成
  time.Sleep(time.Second)
}

func processMsg(msg *wechatbot.Message) {
  content := "<NULL>"
  if msg.Content != "" {
    content = msg.Content
  }
  fmt.Printf("From:%s\nTo:%s\nType:%d\nContent:%s\n\n", msg.FromUserName, msg.ToUserName, msg.Type, content)
  if msg.FromContact == nil || msg.FromContact.Type != wechatbot.ContactFriend {
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
