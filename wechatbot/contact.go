package wechatbot

import (
  "sync"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/conv"
)

const (
  ContactUnknown = iota

  // 好友
  ContactFriend

  // 群聊
  ContactGroup

  // 公众号
  ContactMPS

  // 系统
  ContactSystem
)

var (
  jsonPathNickName   = []string{"NickName"}
  jsonPathRemarkName = []string{"RemarkName"}
  jsonPathUserName   = []string{"UserName"}
  jsonPathVerifyFlag = []string{"VerifyFlag"}
)

type Contact struct {
  NickName   string `json:"nick_name,omitempty"`
  RemarkName string `json:"remark_name,omitempty"`

  // UserName每次登录都不一样，
  // 群聊以@@开头，其他以@开头，内置帐号就直接是名字，如：
  // weixin（微信团队）/filehelper（文件传输助手）/fmessage（朋友消息推荐）
  UserName string `json:"user_name,omitempty"`

  // 联系人类型，
  // 个人和群聊帐号为0，
  // 订阅号为8，
  // 企业号为24（包括扩微信支付），
  // 系统号为56(微信团队官方帐号），
  // 29（未知，招行信用卡为29）
  VerifyFlag int `json:"verify_flag"`

  // 原始数据
  Raw []byte `json:"raw,omitempty"`

  Attr *sync.Map `json:"attr,omitempty"`

  Id string `json:"id,omitempty"`

  // Type是VerifyFlag解析后的值
  Type int `json:"flag"`

  // todo OwnerUin和Bot需要在初始化联系人的时候赋值
  // 联系人所属Bot的uin
  OwnerUin int64 `json:"owner_uin"`

  Bot *Bot `json:"-"`
}

func buildContact(data []byte) *Contact {
  if len(data) == 0 {
    return nil
  }
  ret := &Contact{Raw: data, Attr: &sync.Map{}}
  jsonparser.EachKey(data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      ret.NickName, _ = jsonparser.ParseString(v)
    case 1:
      ret.RemarkName, _ = jsonparser.ParseString(v)
    case 2:
      ret.UserName, _ = jsonparser.ParseString(v)
    case 3:
      vf, _ := jsonparser.ParseInt(v)
      if vf != 0 {
        ret.VerifyFlag = int(vf)
      }
    }
  }, jsonPathNickName, jsonPathRemarkName, jsonPathUserName, jsonPathVerifyFlag)
  if getIdByRemarkName(ret.RemarkName) != 0 {
    ret.Id = ret.RemarkName
  }
  switch ret.VerifyFlag {
  case 0:
    if len(ret.UserName) < 2 {
      ret.Type = ContactUnknown
    } else {
      if (ret.UserName)[0:2] == "@@" {
        ret.Type = ContactGroup
      } else if (ret.UserName)[0:1] == "@" {
        ret.Type = ContactFriend
      } else {
        ret.Type = ContactSystem
      }
    }
  case 8, 24:
    ret.Type = ContactMPS
  case 56:
    ret.Type = ContactSystem
  default:
    ret.Type = ContactUnknown
  }
  return ret
}

func FindContactByUserName(userName string) *Contact {
  if userName == "" {
    return nil
  }
  var ret *Contact
  eachBot(func(b *Bot) bool {
    if b.Contacts != nil {
      if c := b.Contacts.FindByUserName(userName); c != nil {
        ret = c
        return false
      }
    }
    return true
  })
  return ret
}

func (c *Contact) withBot(bot *Bot) {
  if bot == nil {
    return
  }
  c.OwnerUin = bot.req.Uin
  c.Bot = bot
}

func (c *Contact) SendText(text string) error {
  if text == "" {
    return ErrInvalidArgs
  }
  return c.Bot.sendText(c.UserName, text)
}

func (c *Contact) SendImage(data []byte, filename string) (string, error) {
  if len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  return c.Bot.sendMedia(c.UserName, data, filename, MsgImage, sendImageUrlPath)
}

func (c *Contact) SendVideo(data []byte, filename string) (string, error) {
  if len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  return c.Bot.sendMedia(c.UserName, data, filename, MsgVideo, sendVideoUrlPath)
}

func (c *Contact) GetAttrString(attr string, defaultValue string) string {
  if v, ok := c.Attr.Load(attr); ok {
    return conv.String(v, defaultValue)
  }
  return defaultValue
}

func (c *Contact) GetAttrInt(attr string, defaultValue int) int {
  if v, ok := c.Attr.Load(attr); ok {
    return conv.Int(v, defaultValue)
  }
  return defaultValue
}

func (c *Contact) GetAttrInt64(attr string, defaultValue int64) int64 {
  if v, ok := c.Attr.Load(attr); ok {
    return conv.Int64(v, defaultValue)
  }
  return defaultValue
}

func (c *Contact) GetAttrUint(attr string, defaultValue uint) uint {
  if v, ok := c.Attr.Load(attr); ok {
    return conv.Uint(v, defaultValue)
  }
  return defaultValue
}

func (c *Contact) GetAttrUint64(attr string, defaultValue uint64) uint64 {
  if v, ok := c.Attr.Load(attr); ok {
    return conv.Uint64(v, defaultValue)
  }
  return defaultValue
}

func (c *Contact) GetAttrBool(attr string, defaultValue bool) bool {
  if v, ok := c.Attr.Load(attr); ok {
    return conv.Bool(v)
  }
  return defaultValue
}

func (c *Contact) GetAttrBytes(attr string) []byte {
  if v, ok := c.Attr.Load(attr); ok {
    switch ret := v.(type) {
    case []byte:
      return ret
    case string:
      return []byte(ret)
    }
  }
  return nil
}
