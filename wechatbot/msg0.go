package wechatbot

import "github.com/buger/jsonparser"

func PipeMsgHandlers(msg *Message, handlers ...func(*Message)) {
  for _, h := range handlers {
    h(msg)
  }
}

func HandleVerifyMsg(msg *Message) {
  if msg.Type == MsgVerify {
    v, _, _, _ := jsonparser.Get(msg.Raw, "RecommendInfo")
    u, _ := jsonparser.GetString(v, "UserName")
    t, _ := jsonparser.GetString(v, "Ticket")
    if u != "" && t != "" {
      msg.Bot.Accept(u, t)
    }
  }
}

func HandleGroupMsg(msg *Message) {
  if len(msg.Content) >= 39 && msg.Content[33:34] == ":" {
    msg.SpeakerUserName = msg.Content[:33]
    msg.Content = msg.Content[39:]
  }
}
