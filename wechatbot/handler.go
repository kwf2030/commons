package wechatbot

import (
  "github.com/buger/jsonparser"
)

type Handler interface {
  // 登录成功（error == nil），
  // 登录失败（error != nil）
  OnSignIn(*HandlerContext, error)

  // 主动退出（error == nil），
  // 被动退出（error != nil）
  OnSignOut(*HandlerContext, error)

  // 收到二维码（需扫码登录），
  // 第二个参数为二维码的Url
  OnQRCode(*HandlerContext, string)

  // 收到好友申请，
  // 这里的Contact参数只有少量信息，且不在Bot.Contacts内，
  // 第三个参数是用于Bot.Accept的ticket参数，
  // Bot.Accept返回的Contact信息较全，且会自动添加到Bot.Contacts
  OnFriendApply(*HandlerContext, *Contact, string)

  // 好友更新（包括好友资料更新、删除好友或被好友删除），
  // 这里的Contact是已经更新过且在Bot.Contacts中的，
  // 第三个参数暂时无用
  OnFriendUpdate(*HandlerContext, *Contact, int)

  // 加入群聊（包括创建群、被拉入群或加入群），
  // 这里的Contact已经自动添加到Bot.Contacts中了，
  // 第三个参数暂时无用
  OnGroupJoin(*HandlerContext, *Contact, int)

  // 群更新（包括群改名、群成员变更或其他群信息更新），
  // 这里的Contact是已经更新过且在Bot.Contacts中的，
  // 第三个参数暂时无用
  OnGroupUpdate(*HandlerContext, *Contact, int)

  // 退群（包括主动退群或被群主移出群），
  // 这里的Contact已经自动从Bot.Contacts删除了，
  // 第三个参数暂时无用
  OnGroupExit(*HandlerContext, *Contact, int)

  // 收到消息（所有非以上类型的消息）
  OnMessage(*HandlerContext, *Message)
}

type DefaultHandler struct{}

func (h *DefaultHandler) OnSignIn(*HandlerContext, error) {

}

type HandlerContext struct {
  prev, next *HandlerContext
  pipeline   *Pipeline
  handler    Handler
}

func (ctx *HandlerContext) Pipeline() *Pipeline {
  return ctx.pipeline
}

func (ctx *HandlerContext) FireSignIn(e error) {
  if next := ctx.next; next != nil {
    next.handler.OnSignIn(next, e)
  }
}

func (ctx *HandlerContext) FireSignOut(e error) {
  if next := ctx.next; next != nil {
    next.handler.OnSignOut(next, e)
  }
}

func (ctx *HandlerContext) FireQRCode(addr string) {
  if next := ctx.next; next != nil {
    next.handler.OnQRCode(next, addr)
  }
}

func (ctx *HandlerContext) FireFriendApply(c *Contact, ticket string) {
  if next := ctx.next; next != nil {
    next.handler.OnFriendApply(next, c, ticket)
  }
}

func (ctx *HandlerContext) FireFriendUpdate(c *Contact, code int) {
  if next := ctx.next; next != nil {
    next.handler.OnFriendUpdate(next, c, code)
  }
}

func (ctx *HandlerContext) FireGroupJoin(c *Contact, code int) {
  if next := ctx.next; next != nil {
    next.handler.OnGroupJoin(next, c, code)
  }
}

func (ctx *HandlerContext) FireGroupUpdate(c *Contact, code int) {
  if next := ctx.next; next != nil {
    next.handler.OnGroupUpdate(next, c, code)
  }
}

func (ctx *HandlerContext) FireGroupExit(c *Contact, code int) {
  if next := ctx.next; next != nil {
    next.handler.OnGroupExit(next, c, code)
  }
}

func (ctx *HandlerContext) FireMessage(msg *Message) {
  if next := ctx.next; next != nil {
    next.handler.OnMessage(next, msg)
  }
}

func HandleVerifyMsg(msg *Message) *Contact {
  if msg.Type == MsgVerify {
    v, _, _, _ := jsonparser.Get(msg.Raw, "RecommendInfo")
    u, _ := jsonparser.GetString(v, "UserName")
    t, _ := jsonparser.GetString(v, "Ticket")
    if u != "" && t != "" {
      c, _ := msg.Bot.Accept(u, t)
      return c
    }
  }
  return nil
}

func HandleGroupMsg(msg *Message) {
  if len(msg.Content) >= 39 && msg.Content[33:34] == ":" {
    msg.SpeakerUserName = msg.Content[:33]
    msg.Content = msg.Content[39:]
  } else if len(msg.Content) >= 39 && msg.Content[33:34] == ":" {
    msg.SpeakerUserName = msg.Content[:33]
    msg.Content = msg.Content[39:]
  }
}
