package wechatbot

import (
  "io/ioutil"
  "os"
  "strconv"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/times"
)

func (bot *Bot) DownloadQRCode(dst string) (string, error) {
  return bot.req.DownloadQRCode(dst)
}

func (bot *Bot) DownloadAvatar(dst string) (string, error) {
  return bot.req.DownloadAvatar(dst)
}

func (bot *Bot) SendText(toUserName string, text string) error {
  if text == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.Get(toUserName); c != nil {
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
  ret, e := jsonparser.GetInt(resp, "BaseResponse", "Ret")
  if e != nil {
    return e
  }
  if ret != 0 {
    return ErrResp
  }
  return nil
}

func (bot *Bot) SendImage(toUserName string, data []byte, filename string) (string, error) {
  if toUserName == "" || len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return "", ErrInvalidState
  }
  if c := bot.Contacts.Get(toUserName); c != nil {
    return bot.sendMedia(c.UserName, data, filename, MsgImage, sendImageUrlPath)
  }
  return "", ErrContactNotFound
}

func (bot *Bot) SendVideo(toUserName string, data []byte, filename string) (string, error) {
  if toUserName == "" || len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return "", ErrInvalidState
  }
  if c := bot.Contacts.Get(toUserName); c != nil {
    return bot.sendMedia(c.UserName, data, filename, MsgVideo, sendVideoUrlPath)
  }
  return "", ErrContactNotFound
}

func (bot *Bot) sendMedia(toUserName string, data []byte, filename string, msgType int, sendUrlPath string) (string, error) {
  if bot.req == nil {
    return "", ErrInvalidState
  }
  mediaId, e := bot.req.UploadMedia(toUserName, data, filename)
  if e != nil {
    return "", e
  }
  if mediaId == "" {
    return "", ErrResp
  }
  resp, e := bot.req.SendMedia(toUserName, mediaId, msgType, sendUrlPath)
  if e != nil {
    return "", e
  }
  ret, e := jsonparser.GetInt(resp, "BaseResponse", "Ret")
  if e != nil {
    return "", e
  }
  if ret != 0 {
    return "", ErrResp
  }
  return mediaId, nil
}

func (bot *Bot) ForwardImage(toUserName, mediaId string) error {
  if toUserName == "" || mediaId == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.Get(toUserName); c != nil {
    _, e := bot.req.SendMedia(c.UserName, mediaId, MsgImage, sendImageUrlPath)
    return e
  }
  return ErrContactNotFound
}

func (bot *Bot) ForwardVideo(toUserName, mediaId string) error {
  if toUserName == "" || mediaId == "" {
    return ErrInvalidArgs
  }
  if bot.Contacts == nil {
    return ErrInvalidState
  }
  if c := bot.Contacts.Get(toUserName); c != nil {
    _, e := bot.req.SendMedia(c.UserName, mediaId, MsgVideo, sendVideoUrlPath)
    return e
  }
  return ErrContactNotFound
}

// 通过验证且添加到联系人
func (bot *Bot) Accept(c *Contact) (*Contact, error) {
  if c == nil || c.UserName == "" || c.GetAttrString("Ticket", "") == "" {
    return nil, ErrInvalidArgs
  }
  resp, e := bot.req.Verify(c.UserName, c.GetAttrString("Ticket", ""))
  if e != nil {
    return nil, e
  }
  code, e := jsonparser.GetInt(resp, "BaseResponse", "Ret")
  if e != nil {
    return nil, e
  }
  if code != 0 {
    return nil, ErrResp
  }

  resp, e = bot.req.GetContacts(c.UserName)
  if e != nil {
    return nil, e
  }
  code, e = jsonparser.GetInt(resp, "BaseResponse", "Ret")
  if e != nil {
    return nil, e
  }
  if code != 0 {
    return nil, ErrResp
  }
  var ret *Contact
  _, _ = jsonparser.ArrayEach(resp, func(v []byte, _ jsonparser.ValueType, _ int, e error) {
    if e != nil {
      return
    }
    cc := buildContact(v)
    if cc != nil && cc.UserName != "" {
      cc.withBot(bot)
      ret = cc
    }
  }, "ContactList")
  if ret == nil {
    return nil, ErrResp
  }

  bot.Contacts.Add(ret)
  return ret, nil
}

func (bot *Bot) GetAttrString(attr string, defaultValue string) string {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.String(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrInt(attr string, defaultValue int) int {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Int(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrInt64(attr string, defaultValue int64) int64 {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Int64(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrUint(attr string, defaultValue uint) uint {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Uint(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrUint64(attr string, defaultValue uint64) uint64 {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Uint64(v, defaultValue)
  }
  return defaultValue
}

func (bot *Bot) GetAttrBool(attr string, defaultValue bool) bool {
  if v, ok := bot.Attr.Load(attr); ok {
    return conv.Bool(v)
  }
  return defaultValue
}

func deviceId() string {
  return "e" + timestampStringL(15)
}

func timestampString13() string {
  return timestampStringL(13)
}

func timestampString10() string {
  return timestampStringL(10)
}

func timestampStringL(l int) string {
  s := strconv.FormatInt(times.Timestamp(), 10)
  if len(s) > l {
    return s[:l]
  }
  return s
}

func timestampStringR(l int) string {
  s := strconv.FormatInt(times.Timestamp(), 10)
  i := len(s) - l
  if i > 0 {
    return s[i:]
  }
  return s
}

func dumpToFile(filename string, data []byte) {
  if dumpToFileEnabled {
    ioutil.WriteFile(dumpDir+filename, data, os.ModePerm)
  }
}
