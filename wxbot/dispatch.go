package wxbot

import (
  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/pipeline"
)

type verifyMsgHandler struct {
  *Bot
}

func (h *verifyMsgHandler) Handle(ctx *pipeline.HandlerContext, val interface{}) {
  if msg, ok := val.(*Message); ok && msg.Type == MsgVerify {
    v, _, _, _ := jsonparser.Get(msg.raw, "RecommendInfo")
    t, _ := jsonparser.GetString(v, "Ticket")
    c := buildContact(v)
    if c.UserName != "" && t != "" {
      h.handler.OnFriendApply(c, t)
      return
    }
  }
  ctx.Fire(val)
}

type groupMsgHandler struct {
  *Bot
}

func (h *groupMsgHandler) Handle(ctx *pipeline.HandlerContext, val interface{}) {
  if msg, ok := val.(*Message); ok {
    if len(msg.Content) >= 39 && msg.Content[33] == ':' {
      msg.SpeakerUserName = msg.Content[:33]
      msg.Content = msg.Content[39:]
      h.handler.OnMessage(msg, 0)
      return
    } else if len(msg.Content) >= 71 && msg.Content[65] == ':' {
      msg.SpeakerUserName = msg.Content[:33]
      msg.Content = msg.Content[71:]
      h.handler.OnMessage(msg, 0)
      return
    }
  }
  ctx.Fire(val)
}

type dispatchHandler struct {
  *Bot
}

func (h *dispatchHandler) Handle(ctx *pipeline.HandlerContext, val interface{}) {
  switch v := val.(type) {
  case *Message:
    h.handler.OnMessage(v, 0)
  case *Contact:
    if v.Type == ContactFriend {
      h.handler.OnFriendUpdate(v, 0)
    } else if v.Type == ContactGroup {
      h.handler.OnGroupUpdate(v, 0)
    }
  }
}
