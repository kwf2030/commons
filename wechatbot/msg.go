package wechatbot

import (
  "strconv"

  "github.com/buger/jsonparser"
)

const (
  // 自带表情是文本消息，Content字段内容为：[奸笑]，
  // emoji表情也是文本消息，Content字段内容为：<span class="emoji emoji1f633"></span>，
  // 如果连同文字和表情一起发送，Content字段内容是文字和表情直接是混在一起，
  // 位置坐标也是文本消息，Content字段内容为：雨花台区雨花西路(德安花园东):/cgi-bin/mmwebwx-bin/webwxgetpubliclinkimg?url=xxx&msgid=741398718084560243&pictype=location
  MsgText = 1

  // 图片/照片消息
  MsgImage = 3

  // 语音消息
  MsgVoice = 34

  // 被添加好友待验证
  MsgVerify = 37

  MsgFriendRecommend = 40

  // 名片消息
  MsgCard = 42

  // 拍摄（视频消息）
  MsgVideo = 43

  // 动画表情，
  // 包括官方表情包中的表情（Content字段无内容）和自定义的图片表情（Content字段内容为XML）
  MsgAnimEmotion = 47

  MsgLocation = 48

  // 公众号推送的链接，分享的链接（AppMsgType=1/3/5），红包（AppMsgType=2001），
  // 发送的文件，收藏，实时位置共享
  MsgLink = 49

  MsgVoip = 50

  // 登录之后系统发送的初始化消息
  MsgInit = 51

  MsgVoipNotify = 52
  MsgVoipInvite = 53
  MsgVideoCall  = 62

  MsgNotice = 9999

  // 系统消息，
  // 例如通过好友验证，系统会发送"你已添加了..."，"如果陌生人..."，"实时位置共享已结束"的消息，
  MsgSystem = 10000

  // 撤回消息
  MsgRevoke = 10002
)

var (
  jsonPathNewMsgId     = []string{"NewMsgId"}
  jsonPathMsgId        = []string{"MsgId"}
  jsonPathMsgType      = []string{"MsgType"}
  jsonPathContent      = []string{"Content"}
  jsonPathUrl          = []string{"Url"}
  jsonPathFromUserName = []string{"FromUserName"}
  jsonPathToUserName   = []string{"ToUserName"}
  jsonPathCreateTime   = []string{"CreateTime"}
)

type GroupMessage struct {
  FromMemberUserName string
}

type Message struct {
  Id           string
  Type         int
  Content      string
  Url          string
  FromUserName string
  ToUserName   string
  CreateTime   int64
  Raw          []byte

  Bot *Bot
  *GroupMessage
}

func buildMessage(data []byte) *Message {
  if len(data) == 0 {
    return nil
  }
  ret := &Message{Raw: data, GroupMessage: &GroupMessage{}}
  jsonparser.EachKey(data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      id, _ := jsonparser.ParseInt(v)
      if id != 0 {
        ret.Id = strconv.FormatInt(id, 10)
      }
    case 1:
      id, _ := jsonparser.ParseString(v)
      if id != "" && ret.Id == "" {
        ret.Id = id
      }
    case 2:
      t, _ := jsonparser.ParseInt(v)
      if t != 0 {
        ret.Type = int(t)
      }
    case 3:
      ret.Content, _ = jsonparser.ParseString(v)
    case 4:
      ret.Url, _ = jsonparser.ParseString(v)
    case 5:
      ret.FromUserName, _ = jsonparser.ParseString(v)
    case 6:
      ret.ToUserName, _ = jsonparser.ParseString(v)
    case 7:
      ret.CreateTime, _ = jsonparser.ParseInt(v)
    }
  }, jsonPathNewMsgId, jsonPathMsgId, jsonPathMsgType, jsonPathContent, jsonPathUrl, jsonPathFromUserName, jsonPathToUserName, jsonPathCreateTime)
  return ret
}

func (msg *Message) withBot(bot *Bot) {
  if bot == nil {
    return
  }
  msg.Bot = bot
}

func (msg *Message) ReplyText(text string) error {
  if text == "" {
    return ErrInvalidArgs
  }
  return msg.Bot.sendText(msg.FromUserName, text)
}

func (msg *Message) ReplyImage(data []byte, filename string) (string, error) {
  if len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  return msg.Bot.sendMedia(msg.FromUserName, data, filename, MsgImage, sendImageUrlPath)
}

func (msg *Message) ReplyVideo(data []byte, filename string) (string, error) {
  if len(data) == 0 || filename == "" {
    return "", ErrInvalidArgs
  }
  return msg.Bot.sendMedia(msg.FromUserName, data, filename, MsgVideo, sendVideoUrlPath)
}
