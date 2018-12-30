package wechatbot

import (
  "strconv"

  "github.com/buger/jsonparser"
)

const (
  jsonPathBaseResponse = "BaseResponse"
  jsonPathRet          = "Ret"
  jsonPathContactList  = "ContactList"
)

func (bot *Bot) DownloadQRCode(dst string) (string, error) {
  return bot.req.DownloadQRCode(dst)
}

func (bot *Bot) DownloadAvatar(dst string) (string, error) {
  return bot.req.DownloadAvatar(dst)
}

func (bot *Bot) SendTextToUserID(id string, text string) error {
  if text == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.FindByID(id); c != nil {
    return bot.sendText(c.UserName, text)
  }
  return ErrContactNotFound
}

func (bot *Bot) SendTextToUserName(toUserName string, text string) error {
  if text == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.FindByUserName(toUserName); c != nil {
    return bot.sendText(c.UserName, text)
  }
  return ErrContactNotFound
}

func (bot *Bot) sendText(toUserName string, text string) error {
  if bot.req == nil {
    return ErrInvalidState
  }
  resp, e := bot.req.SendText(toUserName, text)
  if e != nil {
    return e
  }
  ret, e := jsonparser.GetInt(resp, jsonPathBaseResponse, jsonPathRet)
  if e != nil {
    return e
  }
  if ret != 0 {
    return ErrResp
  }
  return nil
}

func (bot *Bot) SendImageToUserID(id string, data []byte, filename string) (string, error) {
  if id == "" || len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return "", ErrInvalidState
  }
  if c := bot.Contacts.FindByID(id); c != nil {
    return bot.sendMedia(c.UserName, data, filename, MsgImage, sendImageUrlPath)
  }
  return "", ErrContactNotFound
}

func (bot *Bot) SendImageToUserName(toUserName string, data []byte, filename string) (string, error) {
  if toUserName == "" || len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return "", ErrInvalidState
  }
  if c := bot.Contacts.FindByUserName(toUserName); c != nil {
    return bot.sendMedia(c.UserName, data, filename, MsgImage, sendImageUrlPath)
  }
  return "", ErrContactNotFound
}

func (bot *Bot) SendVideoToUserID(id string, data []byte, filename string) (string, error) {
  if id == "" || len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return "", ErrInvalidState
  }
  if c := bot.Contacts.FindByID(id); c != nil {
    return bot.sendMedia(c.UserName, data, filename, MsgVideo, sendVideoUrlPath)
  }
  return "", ErrContactNotFound
}

func (bot *Bot) SendVideoToUserName(toUserName string, data []byte, filename string) (string, error) {
  if toUserName == "" || len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return "", ErrInvalidState
  }
  if c := bot.Contacts.FindByUserName(toUserName); c != nil {
    return bot.sendMedia(c.UserName, data, filename, MsgVideo, sendVideoUrlPath)
  }
  return "", ErrContactNotFound
}

func (bot *Bot) sendMedia(toUserName string, data []byte, filename string, msgType int, sendUrlPath string) (string, error) {
  if bot.req == nil {
    return "", ErrInvalidState
  }
  mediaID, e := bot.req.UploadMedia(toUserName, data, filename)
  if e != nil {
    return "", e
  }
  if mediaID == "" {
    return "", ErrResp
  }
  resp, e := bot.req.SendMedia(toUserName, mediaID, msgType, sendUrlPath)
  if e != nil {
    return "", e
  }
  ret, e := jsonparser.GetInt(resp, jsonPathBaseResponse, jsonPathRet)
  if e != nil {
    return "", e
  }
  if ret != 0 {
    return "", ErrResp
  }
  return mediaID, nil
}

func (bot *Bot) ForwardImageToUserID(id, mediaID string) error {
  if id == "" || mediaID == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.FindByID(id); c != nil {
    _, e := bot.req.SendMedia(c.UserName, mediaID, MsgImage, sendImageUrlPath)
    return e
  }
  return ErrContactNotFound
}

func (bot *Bot) ForwardImageToUserName(toUserName, mediaID string) error {
  if toUserName == "" || mediaID == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.FindByUserName(toUserName); c != nil {
    _, e := bot.req.SendMedia(c.UserName, mediaID, MsgImage, sendImageUrlPath)
    return e
  }
  return ErrContactNotFound
}

func (bot *Bot) ForwardVideoToUserID(id, mediaID string) error {
  if id == "" || mediaID == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.FindByID(id); c != nil {
    _, e := bot.req.SendMedia(c.UserName, mediaID, MsgVideo, sendVideoUrlPath)
    return e
  }
  return ErrContactNotFound
}

func (bot *Bot) ForwardVideoToUserName(toUserName, mediaID string) error {
  if toUserName == "" || mediaID == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.FindByUserName(toUserName); c != nil {
    _, e := bot.req.SendMedia(c.UserName, mediaID, MsgVideo, sendVideoUrlPath)
    return e
  }
  return ErrContactNotFound
}

// VerifyAndRemark封装了Verify、GetContacts和Remark三个请求，
// GetContact成功后会设置ID并添加到本地联系人中（如果开启持久化功能的话），
// 之后再Remark，如果Remark失败，不会影响联系人数据，
// 但是在下次微信登录后发现联系人没有Remark会再次Remark，ID可能会跟这次不一样
func (bot *Bot) VerifyAndRemark(toUserName, ticket string) (*Contact, error) {
  if toUserName == "" || ticket == "" {
    return nil, ErrInvalidArgs
  }
  resp, e := bot.req.Verify(toUserName, ticket)
  if e != nil {
    return nil, ErrReq
  }
  ret, e := jsonparser.GetInt(resp, jsonPathBaseResponse, jsonPathRet)
  if e != nil {
    return nil, e
  }
  if ret != 0 {
    return nil, ErrResp
  }

  resp, e = bot.req.GetContacts(toUserName)
  if e != nil {
    return nil, ErrReq
  }
  ret, e = jsonparser.GetInt(resp, jsonPathBaseResponse, jsonPathRet)
  if e != nil {
    return nil, e
  }
  if ret != 0 {
    return nil, ErrResp
  }
  var contact *Contact
  _, _ = jsonparser.ArrayEach(resp, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    c := buildContact(v)
    if c != nil && c.UserName != "" {
      contact = c
    }
  }, jsonPathContactList)
  // todo 给contact的字段赋值

  if b, ok := bot.Attr.Load(attrIdEnabled); !ok || !b.(bool) {
    bot.Contacts.Add(contact)
    return contact, nil
  }

  id := bot.Contacts.nextID()
  contact.id = id
  contact.Id = strconv.FormatUint(id, 10)
  bot.Contacts.Add(contact)
  resp, e = bot.req.Remark(toUserName, contact.Id)
  if e != nil {
    return nil, ErrReq
  }
  ret, e = jsonparser.GetInt(resp, jsonPathRet)
  if e != nil {
    return nil, e
  }
  if ret != 0 {
    return nil, ErrResp
  }
  return contact, nil
}
