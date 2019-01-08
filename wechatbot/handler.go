package wechatbot

import (
  "github.com/buger/jsonparser"
)

var (
  jsonPathAddMsgList     = []string{"AddMsgList"}
  jsonPathDelContactList = []string{"DelContactList"}
  jsonPathModContactList = []string{"ModContactList"}
  jsonPathSyncCheckKey   = []string{"SyncCheckKey"}
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
  // 这里的Contact只有UserName和NickName，且不在Bot.Contacts内，
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

  // 收到消息，
  // 第三个参数暂时无用
  OnMessage(*HandlerContext, *Message, int)

  // 用于内部数据分发和回调，
  // 不要实现这个方法，因为不会被调用（除非自己调用HandlerContext.FireData），
  // 第三和四个参数分别表示syncCheck的code和selector，
  OnData(*HandlerContext, []byte, int, int)
}

type HandlerContext struct {
  prev, next *HandlerContext
  pipeline   *Pipeline
  handler    Handler
}

func (ctx *HandlerContext) Bot() *Bot {
  return ctx.pipeline.bot
}

func (ctx *HandlerContext) Pipeline() *Pipeline {
  return ctx.pipeline
}

func (ctx *HandlerContext) Prev() *HandlerContext {
  ctx.pipeline.m.RLock()
  ret := ctx.prev
  if ret == ctx.pipeline.head {
    ret = nil
  }
  ctx.pipeline.m.RUnlock()
  return ret
}

func (ctx *HandlerContext) Next() *HandlerContext {
  ctx.pipeline.m.RLock()
  ret := ctx.next
  if ret == ctx.pipeline.tail {
    ret = nil
  }
  ctx.pipeline.m.RUnlock()
  return ret
}

func (ctx *HandlerContext) FireSignIn(e error) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnSignIn(next, e)
  }
}

func (ctx *HandlerContext) FireSignOut(e error) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnSignOut(next, e)
  }
}

func (ctx *HandlerContext) FireQRCode(addr string) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnQRCode(next, addr)
  }
}

func (ctx *HandlerContext) FireFriendApply(c *Contact, ticket string) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnFriendApply(next, c, ticket)
  }
}

func (ctx *HandlerContext) FireFriendUpdate(c *Contact, code int) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnFriendUpdate(next, c, code)
  }
}

func (ctx *HandlerContext) FireGroupJoin(c *Contact, code int) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnGroupJoin(next, c, code)
  }
}

func (ctx *HandlerContext) FireGroupUpdate(c *Contact, code int) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnGroupUpdate(next, c, code)
  }
}

func (ctx *HandlerContext) FireGroupExit(c *Contact, code int) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnGroupExit(next, c, code)
  }
}

func (ctx *HandlerContext) FireMessage(msg *Message, code int) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnMessage(next, msg, code)
  }
}

func (ctx *HandlerContext) FireData(data []byte, code, selector int) {
  ctx.pipeline.m.RLock()
  next := ctx.next
  ctx.pipeline.m.RUnlock()
  if next != nil {
    next.handler.OnData(next, data, code, selector)
  }
}

type DefaultHandler struct{}

func (h DefaultHandler) OnSignIn(ctx *HandlerContext, e error) {
  ctx.FireSignIn(e)
}

func (h DefaultHandler) OnSignOut(ctx *HandlerContext, e error) {
  ctx.FireSignOut(e)
}

func (h DefaultHandler) OnQRCode(ctx *HandlerContext, addr string) {
  ctx.FireQRCode(addr)
}

func (h DefaultHandler) OnFriendApply(ctx *HandlerContext, c *Contact, ticket string) {
  ctx.FireFriendApply(c, ticket)
}

func (h DefaultHandler) OnFriendUpdate(ctx *HandlerContext, c *Contact, code int) {
  ctx.FireFriendUpdate(c, code)
}

func (h DefaultHandler) OnGroupJoin(ctx *HandlerContext, c *Contact, code int) {
  ctx.FireGroupJoin(c, code)
}

func (h DefaultHandler) OnGroupUpdate(ctx *HandlerContext, c *Contact, code int) {
  ctx.FireGroupUpdate(c, code)
}

func (h DefaultHandler) OnGroupExit(ctx *HandlerContext, c *Contact, code int) {
  ctx.FireGroupExit(c, code)
}

func (h DefaultHandler) OnMessage(ctx *HandlerContext, msg *Message, code int) {
  ctx.FireMessage(msg, code)
}

func (h DefaultHandler) OnData(ctx *HandlerContext, data []byte, code, selector int) {
  ctx.FireData(data, code, selector)
}

type DispatchHandler struct {
  DefaultHandler
}

func (h *DispatchHandler) OnData(ctx *HandlerContext, data []byte, code, selector int) {
  var addMsgList []*Message
  var delContactList, modContactList []*Contact
  jsonparser.EachKey(data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      addMsgList = h.parseMsgList(v, ctx.Bot())
    case 1:
      delContactList = h.parseContactList(v, ctx.Bot())
    case 2:
      modContactList = h.parseContactList(v, ctx.Bot())
    case 3:
      b, _, _, e := jsonparser.Get(data, "SyncKey")
      if e == nil {
        sk := parseSyncKey(b)
        if sk != nil && sk.Count > 0 {
          ctx.Bot().req.SyncKeys = sk
        }
      }
    }
  }, jsonPathAddMsgList, jsonPathDelContactList, jsonPathModContactList, jsonPathSyncCheckKey)
  // 没开启验证如果被添加好友，
  // ModContactList（对方信息）和AddMsgList（添加到通讯录的系统提示）会一起收到，
  // 所以要先处理完Contact后再处理Message（避免找不到发送者），
  // 虽然之后也能一直收到此人的消息，但要想主动发消息，仍需要手动添加好友，
  // 不添加的话下次登录时好友列表中也没有此人，
  // 目前Web微信好像没有添加好友的功能，所以只能开启验证（通过验证即可添加好友）
  for _, c := range modContactList {
    ctx.Bot().Contacts.Add(c)
    if c.Type == ContactFriend {
      ctx.FireFriendUpdate(c, 0)
    } else if c.Type == ContactGroup {
      ctx.FireGroupUpdate(c, 0)
    }
  }
  for _, c := range delContactList {
    ctx.Bot().Contacts.Remove(c.UserName)
    if c.Type == ContactFriend {
      ctx.FireFriendUpdate(c, 0)
    } else if c.Type == ContactGroup {
      ctx.FireGroupUpdate(c, 0)
    }
  }
  for _, m := range addMsgList {
    ctx.FireMessage(m, 0)
  }
}

func (h *DispatchHandler) parseContactList(data []byte, bot *Bot) []*Contact {
  ret := make([]*Contact, 0, 2)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    userName, _ := jsonparser.GetString(v, "UserName")
    if userName == "" {
      return
    }
    c := bot.Contacts.Get(userName)
    if c == nil {
      c = buildContact(v)
      c.withBot(bot)
    }
    ret = append(ret, c)
  })
  return ret
}

func (h *DispatchHandler) parseMsgList(data []byte, bot *Bot) []*Message {
  ret := make([]*Message, 0, 2)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    msg := buildMessage(v)
    if msg != nil && msg.Id != "" {
      msg.withBot(bot)
      ret = append(ret, msg)
    }
  })
  return ret
}

type VerifyMsgHandler struct {
  DefaultHandler
}

func (h *VerifyMsgHandler) OnMessage(ctx *HandlerContext, msg *Message, code int) {
  if msg.Type == MsgVerify {
    v, _, _, _ := jsonparser.Get(msg.Raw, "RecommendInfo")
    t, _ := jsonparser.GetString(v, "Ticket")
    c := buildContact(v)
    if c.UserName != "" && t != "" {
      ctx.FireFriendApply(c, t)
      return
    }
  }
  ctx.FireMessage(msg, code)
}

type GroupMsgHandler struct {
  DefaultHandler
}

func (h *GroupMsgHandler) OnMessage(ctx *HandlerContext, msg *Message, code int) {
  if len(msg.Content) >= 39 && msg.Content[33] == ':' {
    msg.SpeakerUserName = msg.Content[:33]
    msg.Content = msg.Content[39:]
  } else if len(msg.Content) >= 71 && msg.Content[65] == ':' {
    msg.SpeakerUserName = msg.Content[:33]
    msg.Content = msg.Content[71:]
  }
  ctx.FireMessage(msg, code)
}
