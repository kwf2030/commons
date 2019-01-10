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
    h.handler.OnContact(v, 0)
  }
}
