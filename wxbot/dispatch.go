package wxbot

import (
  "github.com/buger/jsonparser"
)

func (bot *Bot) dispatch(syncCheck syncCheckResp, data []byte) {
  var addMsgList []*Message
  var delContactList, modContactList []*Contact
  jsonparser.EachKey(data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      addMsgList = bot.parseSyncMsgList(v)
    case 1:
      delContactList = bot.parseSyncContactList(v)
    case 2:
      modContactList = bot.parseSyncContactList(v)
    case 3:
      sk := parseSyncKey(v)
      if sk.Count > 0 {
        bot.session.SyncKey = sk
      }
    }
  }, jsonPathAddMsgList, jsonPathDelContactList, jsonPathModContactList, jsonPathSyncCheckKey)
  for _, c := range modContactList {
    r.syncPipeline.Fire(c)
  }
  for _, c := range delContactList {
    r.syncPipeline.Fire(c)
  }
  for _, m := range addMsgList {
    if ok := bot.processVerifyMsg(m); ok {
      continue
    }
    if ok := bot.processGroupMsg(m); ok {
      continue
    }
    bot.handler.OnMessage(m, 0)
  }
}

func (bot *Bot) parseSyncContactList(data []byte) []*Contact {
  ret := make([]*Contact, 0, 2)
  _, _ = jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    userName, _ := jsonparser.GetString(v, "UserName")
    if userName == "" {
      return
    }
    c := buildContact(v)
    if c != nil && c.UserName != "" {
      c.withBot(bot)
      ret = append(ret, c)
    }
  })
  return ret
}

func (bot *Bot) parseSyncMsgList(data []byte) []*Message {
  ret := make([]*Message, 0, 2)
  jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
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

func (bot *Bot) processVerifyMsg(msg *Message) bool {
  if msg.Type == MsgVerify {
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

func (bot *Bot) processGroupMsg(msg *Message) bool {
  if len(msg.Content) >= 39 && msg.Content[33] == ':' {
    msg.SpeakerUserName = msg.Content[:33]
    msg.Content = msg.Content[39:]
  } else if len(msg.Content) >= 71 && msg.Content[65] == ':' {
    msg.SpeakerUserName = msg.Content[:33]
    msg.Content = msg.Content[71:]
  }
}
