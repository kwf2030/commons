package wxbot

type Handler interface {
  // 登录成功（error == nil），
  // 登录失败（error != nil）
  OnSignIn(error)

  // 主动退出（error == nil），
  // 被动退出（error != nil）
  OnSignOut(error)

  // 收到二维码（需扫码登录），
  // 第二个参数为二维码的Url
  OnQRCode(string)

  // 收到好友申请，
  // 这里的Contact只有UserName和NickName，且不在Bot.Contacts内，
  // 第三个参数是用于Bot.Accept的ticket参数，
  // Bot.Accept返回的Contact信息较全，且会自动添加到Bot.Contacts
  OnFriendApply(*Contact, string)

  // 好友更新（包括好友资料更新、删除好友或被好友删除），
  // 第二个参数暂时无用
  OnFriendUpdate(*Contact, int)

  // 加入群聊（包括创建群、被拉入群或加入群），
  // 第二个参数暂时无用
  OnGroupJoin(*Contact, int)

  // 群更新（包括群改名、群成员变更或其他群信息更新），
  // 第二个参数暂时无用
  OnGroupUpdate(*Contact, int)

  // 退群（包括主动退群或被群主移出群），
  // 第二个参数暂时无用
  OnGroupExit(*Contact, int)

  // 收到消息，
  // 第二个参数暂时无用
  OnMessage(*Message, int)
}

type DefaultHandler struct{}

func (*DefaultHandler) OnSignIn(error) {}

func (*DefaultHandler) OnSignOut(error) {}

func (*DefaultHandler) OnQRCode(string) {}

func (*DefaultHandler) OnFriendApply(*Contact, string) {}

func (*DefaultHandler) OnFriendUpdate(*Contact, int) {}

func (*DefaultHandler) OnGroupJoin(*Contact, int) {}

func (*DefaultHandler) OnGroupUpdate(*Contact, int) {}

func (*DefaultHandler) OnGroupExit(*Contact, int) {}

func (*DefaultHandler) OnMessage(*Message, int) {}
