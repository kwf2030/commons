package wechatbot

import (
  "strconv"

  "github.com/buger/jsonparser"
)

const (
  // 自带表情是文本消息，Content字段内容为：[奸笑]，
  // emoji表情也是文本消息，Content字段内容为：<span class="emoji emoji1f633"></span>，
  // 如果连同文字和表情一起发送，Content字段内容是文字和表情直接是混在一起，
  // 位置坐标也是文本消息，Content字段内容为：
  // 雨花台区雨花西路(德安花园东):/cgi-bin/mmwebwx-bin/webwxgetpubliclinkimg?url=xxx&msgid=741398718084560243&pictype=location
  MsgText = 1

  // 图片/照片消息，
  // Content字段内容为XML，Content字段内容为：
  // <?xml version="1.0"?>
  // <msg>
  // <img aeskey="" encryver="" cdnthumbaeskey="" cdnthumburl="" cdnthumblength=""
  //   cdnthumbheight="" cdnthumbwidth="" cdnmidheight="" cdnmidwidth="" cdnhdheight=""
  //   cdnhdwidth="" cdnmidimgurl="" length="" md5="" /><br/>
  // </msg>
  MsgImage = 3

  MsgVoice = 34

  // 被添加好友待验证，Content内容为：
  // <msg fromusername="kwf2030" encryptusername="v1_400be59c1cd145d71bcd4a389b68456833bfdba992d524a563494beed6def517@stranger"
  // fromnickname="kwf2030" content="我是客户"  shortpy="KWF2030" imagestatus="3" scene="30"
  // country="CN" province="Jiangsu" city="Nanjing" sign="" percard="1" sex="1" alias="" weibo=""
  // weibonickname="" albumflag="0" albumstyle="0" albumbgimgid="" snsflag="17" snsbgimgid=""
  // snsbgobjectid="0" mhash="21b5f7503b9728a74f69ffe2ac4a81b8"
  // mfullhash="21b5f7503b9728a74f69ffe2ac4a81b8"
  // bigheadimgurl="http://wx.qlogo.cn/mmhead/ver_1/as2mDdcIHonnibUkbSzmyAZ4eRPFv67M7IOLXhE4ULXQaRESLaNnLlsjHGvFuNXicnqYmxCXCZFjziaGQetfFyRhQ/0"
  // smallheadimgurl="http://wx.qlogo.cn/mmhead/ver_1/as2mDdcIHonnibUkbSzmyAZ4eRPFv67M7IOLXhE4ULXQaRESLaNnLlsjHGvFuNXicnqYmxCXCZFjziaGQetfFyRhQ/96"
  // ticket="v2_1604664f28c4e339b63f5299ef578d15350d9b02ee5b8137b0c568f5423fa5adfe843d9a7478dbf21395f26ae4567896f52e6cdd9f2971b81f06332c1f2c91bf@stranger"
  // opcode="2" googlecontact="" qrticket="" chatroomusername="" sourceusername="" sourcenickname="">
  // <brandlist count="0" ver="683212005"></brandlist>
  // </msg>
  MsgVerify = 37

  MsgFriendRecommend = 40

  // 名片消息，Content字段内容为：
  // <?xml version="1.0"?>
  // <msg bigheadimgurl="http://xxx" smallheadimgurl="http://xxx" username="v1_xxx@stranger" nickname=""
  // shortpy="" alias="" imagestatus="" scene="" province="" city="" sign="" sex="" certflag=""
  // certinfo="" brandIconUrl="" brandHomeUrl="" brandSubscriptConfigUrl="" brandFlags=""
  // regionCode="" antispamticket="v2_xxx@stranger" />
  MsgCard = 42

  // 拍摄（视频消息）
  MsgVideo = 43

  // 动画表情，
  // 包括官方表情包中的表情（Content字段无内容）和自定义的图片表情（Content字段内容为XML）
  MsgAnimEmotion = 47

  MsgLocation = 48

  // 公众号推送的链接，
  // 发送的文件也是链接消息，
  // 分享的链接（AppMsgType=1/3/5），
  // 红包（AppMsgType=2001）,
  // 收藏也是连接消息，
  // 实时位置共享也是链接消息，Content字段内容为：
  // <msg>
  // <appmsg appid="" sdkver="0">
  // <type>17</type>
  // <title><![CDATA[我发起了位置共享]]></title>
  // </appmsg>
  // <fromusername>kwf2030</fromusername>
  // </msg>
  MsgLink = 49

  MsgVoip = 50

  // 登录之后系统发送的初始化消息
  MsgInit = 51

  MsgVoipNotify = 52
  MsgVoipInvite = 53
  MsgVideoCall  = 62

  MsgNotice = 9999

  // 系统消息，
  // 例如通过好友验证，系统会发送"你已添加了..."和"如果陌生人..."的消息，
  // 例如"实时位置共享已结束"的消息
  MsgSystem = 10000

  // 撤回消息，Content字段内容为：
  // <sysmsg type="revokemsg">
  // <revokemsg>
  // <session>Nickname</session>
  // <oldmsgid>1057920614</oldmsgid>
  // <msgid>2360839023010332147</msgid>
  // <replacemsg><![CDATA["Nickname" 撤回了一条消息]]></replacemsg>
  // </revokemsg>
  // </sysmsg>
  MsgRevoke = 10002
)

var (
  jsonPathContent      = []string{"Content"}
  jsonPathCreateTime   = []string{"CreateTime"}
  jsonPathFromUserName = []string{"FromUserName"}
  jsonPathToUserName   = []string{"ToUserName"}
  jsonPathMsgId        = []string{"MsgId"}
  jsonPathNewMsgId     = []string{"NewMsgId"}
  jsonPathMsgType      = []string{"MsgType"}
  jsonPathUrl          = []string{"Url"}
)

type Message struct {
  Content      string `json:"content,omitempty"`
  CreateTime   int64  `json:"create_time"`
  FromUserName string `json:"from_user_name,omitempty"`
  ToUserName   string `json:"to_user_name,omitempty"`
  Id           string `json:"id,omitempty"`
  Type         int    `json:"type"`
  Url          string `json:"url,omitempty"`
  Raw          []byte `json:"raw,omitempty"`

  FromUserId  string   `json:"from_user_id,omitempty"`
  ToUserId    string   `json:"to_user_id,omitempty"`
  FromContact *Contact `json:"-"`
  ToContact   *Contact `json:"-"`
  Bot         *Bot     `json:"-"`
}

func buildMessage(data []byte) *Message {
  if len(data) == 0 {
    return nil
  }
  ret := &Message{Raw: data}
  jsonparser.EachKey(data, func(i int, v []byte, _ jsonparser.ValueType, e error) {
    if e != nil {
      return
    }
    switch i {
    case 0:
      ret.Content, _ = jsonparser.ParseString(v)
    case 1:
      ret.CreateTime, _ = jsonparser.ParseInt(v)
    case 2:
      ret.FromUserName, _ = jsonparser.ParseString(v)
    case 3:
      ret.ToUserName, _ = jsonparser.ParseString(v)
    case 4:
      id, _ := jsonparser.ParseString(v)
      if id != "" && ret.Id == "" {
        ret.Id = id
      }
    case 5:
      id, _ := jsonparser.ParseInt(v)
      if id != 0 {
        ret.Id = strconv.FormatInt(id, 10)
      }
    case 6:
      t, _ := jsonparser.ParseInt(v)
      if t != 0 {
        ret.Type = int(t)
      }
    case 7:
      ret.Url, _ = jsonparser.ParseString(v)
    }
  }, jsonPathContent, jsonPathCreateTime, jsonPathFromUserName, jsonPathToUserName, jsonPathMsgId, jsonPathNewMsgId, jsonPathMsgType, jsonPathUrl)
  return ret
}

func (msg *Message) withBot(bot *Bot) {
  if bot == nil {
    return
  }
  if c := bot.Contacts.FindByUserName(msg.FromUserName); c != nil {
    msg.FromUserId = c.Id
    msg.FromContact = c
  }
  if c := bot.Contacts.FindByUserName(msg.ToUserName); c != nil {
    msg.ToUserId = c.Id
    msg.ToContact = c
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
