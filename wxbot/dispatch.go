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
    u, _ := jsonparser.GetString(msg.raw, "RecommendInfo", "UserName")
    t, _ := jsonparser.GetString(msg.raw, "RecommendInfo", "Ticket")
    if u != "" && t != "" {
      c, _ := h.Accept(u, t)
      if c != nil {
        h.handler.OnContact(c, 0)
        return
      }
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
    } else if len(msg.Content) >= 71 && msg.Content[65] == ':' {
      msg.SpeakerUserName = msg.Content[:33]
      msg.Content = msg.Content[71:]
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
    if v.FromUserName != h.session.UserName {
      h.updateContact(v.FromUserName)
    }
    if v.ToUserName != h.session.UserName {
      h.updateContact(v.ToUserName)
    }
    h.handler.OnMessage(v, 0)
  case *Contact:
    if c, b := h.updateContact(v.UserName); b {
      h.handler.OnContact(c, 0)
    } else {
      h.handler.OnContact(v, 0)
    }
  }
}

func (h *dispatchHandler) updateContact(userName string) (*Contact, bool) {
  c := h.contacts.Get(userName)
  b := false
  if c == nil {
    c, _ = h.GetContactFromServer(userName)
    if c != nil && c.UserName != "" {
      h.contacts.Add(c)
      b = true
    }
  }
  return c, b
}
