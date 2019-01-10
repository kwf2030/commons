package wxbot

import (
  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/times"
)

var (
  jsonPathAddMsgList     = []string{"AddMsgList"}
  jsonPathDelContactList = []string{"DelContactList"}
  jsonPathModContactList = []string{"ModContactList"}
  jsonPathSyncCheckKey   = []string{"SyncCheckKey"}
)

type dispatchHandler struct {
  *Bot
}

func (h *dispatchHandler) Handle(ctx *handlerCtx, evt event) {
  if evt.err != nil && evt.what != eventSync && evt.what != eventSignOut {
    h.callback.OnSignIn(evt.err)
    return
  }
  switch evt.what {
  case eventSync:
    h.handleSync(ctx, evt)
  case eventQR:
    h.callback.OnQRCode(h.session.QRCodeUrl)
  case eventRedirect:
    h.updatePaths()
  case eventInit:
    h.self = evt.val.(*Contact)
  case eventContacts:
    h.contacts = initContacts(evt.val.([]*Contact), h.Bot)
    h.startTime = times.Now()
    h.session.State = StateRunning
    botsMutex.Lock()
    bots[h.session.Uin] = h.Bot
    botsMutex.Unlock()
    h.callback.OnSignIn(nil)
  case eventSignOut:
    h.Stop()
    var e error
    syncCheck := evt.val.(syncCheckResp)
    if syncCheck.code != 1101 {
      e = ErrSignOut
    }
    h.callback.OnSignOut(e)
  }
}

func (h *dispatchHandler) handleSync(ctx *handlerCtx, evt event) {
  var addMsgList []*Message
  var delContactList, modContactList []*Contact
  jsonparser.EachKey(evt.data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      addMsgList = h.parseMsgList(v)
    case 1:
      delContactList = h.parseContactList(v)
    case 2:
      modContactList = h.parseContactList(v)
    case 3:
      sk := parseSyncKey(v)
      if sk.Count > 0 {
        h.session.SyncKey = sk
      }
    }
  }, jsonPathAddMsgList, jsonPathDelContactList, jsonPathModContactList, jsonPathSyncCheckKey)
  for _, c := range modContactList {
    // h.contacts.Add(c)
    if c.Type == ContactFriend {
      h.callback.OnFriendUpdate(c, 0)
    } else if c.Type == ContactGroup {
      h.callback.OnGroupUpdate(c, 0)
    }
  }
  for _, c := range delContactList {
    h.contacts.Remove(c.UserName)
    if c.Type == ContactFriend {
      h.callback.OnFriendUpdate(c, 1)
    } else if c.Type == ContactGroup {
      h.callback.OnGroupUpdate(c, 1)
    }
  }
  for _, m := range addMsgList {
    h.callback.OnMessage(m, 0)
  }
}

func (h *dispatchHandler) parseContactList(data []byte) []*Contact {
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
      c.withBot(h.Bot)
      ret = append(ret, c)
    }
  })
  return ret
}

func (h *dispatchHandler) parseMsgList(data []byte) []*Message {
  ret := make([]*Message, 0, 2)
  jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    msg := buildMessage(v)
    if msg != nil && msg.Id != "" {
      msg.withBot(h.Bot)
      ret = append(ret, msg)
    }
  })
  return ret
}

type verifyMsgHandler struct {
  *Bot
}

func (h *verifyMsgHandler) Handle(ctx *handlerCtx, evt event) {
  if msg, ok := evt.val.(*Message); ok && msg.Type == MsgVerify {
    v, _, _, _ := jsonparser.Get(msg.raw, "RecommendInfo")
    t, _ := jsonparser.GetString(v, "Ticket")
    c := buildContact(v)
    if c.UserName != "" && t != "" {
      h.callback.OnFriendApply(c, t)
      return
    }
  }
  ctx.Fire(evt)
}

type groupMsgHandler struct {
  *Bot
}

func (h *groupMsgHandler) Handle(ctx *handlerCtx, evt event) {
  if msg, ok := evt.val.(*Message); ok {
    if len(msg.Content) >= 39 && msg.Content[33] == ':' {
      msg.SpeakerUserName = msg.Content[:33]
      msg.Content = msg.Content[39:]
    } else if len(msg.Content) >= 71 && msg.Content[65] == ':' {
      msg.SpeakerUserName = msg.Content[:33]
      msg.Content = msg.Content[71:]
    }
  }
  ctx.Fire(evt)
}
