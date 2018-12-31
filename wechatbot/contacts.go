package wechatbot

import (
  "strconv"
  "strings"
  "sync"
)

const (
  idInitial       = uint64(2018E4)
  idGeneralOffset = uint64(300)
  idGlobalOffset  = uint64(1E7)
  idSelfOffset    = uint64(1E6)
)

// 表示一些内置账号的Id的偏移量，
// UserName=>Id
var internalIds = map[string]uint64{
  "weixin":     1,
  "filehelper": 2,
  "fmessage":   3,
}

type Contacts struct {
  mu *sync.RWMutex

  // UserName=>*Contact
  userNameMap map[string]*Contact
  // Id=>*Contact
  idMap map[string]*Contact

  // 当前所有联系人最大的id，用于生成下一次的联系人id，
  // 每次启动时会遍历所有id，找出最大的赋值给maxId，
  // 后面若要再生成Id，用此值自增转为string即可
  maxId uint64

  Bot *Bot
}

/*func initContacts(data []*Contact, bot *Bot) *Contacts {
  m := map[int]string{}
  m[2] = "22"
  ret := &Contacts{
    Bot:       bot,
    contacts:  sync.Map{},
    userNames: sync.Map{},
  }
  if len(data) == 0 {
    return ret
  }
  if !bot.Attr[AttrPersistentIDEnabled].(bool) {
    for _, c := range data {
      c.Bot = bot
      ret.contacts.Store(c.UserName, c)
    }
    return ret
  }
  // 第一次循环，处理已备注的联系人
  for _, v := range data {
    if v.Flag != ContactFriend {
      continue
    }
    v.Bot = bot
    if remark := conv.GetString(v.Raw, "RemarkName", ""); remark != "" {
      if id := parseRemarkToID(remark); id != 0 {
        if ret.maxID < id {
          ret.maxID = id
        }
        v.ID = strconv.FormatUint(id, 10)
        ret.contacts.Store(v.UserName, v)
        ret.userNames.Store(v.ID, v.UserName)
      }
    }
  }
  if ret.maxID == 0 {
    // 如果Bot从未设置过联系人的ID（备注），那么起始ID就是根据当前已经运行的Bot的个数来决定的，
    // 这是为了当有多个Bot的时候，每个Bot的联系人ID是唯一的且不会重复
    l := uint64(CountBots() - 1)
    ret.maxID = idInitial + (l * idGlobalOffset) + idGeneralOffset
  }
  initial := ret.initialID()
  // 第二次循环，处理其他联系人
  for _, v := range data {
    if (v.Flag != ContactFriend && v.Flag != ContactSystem) || v.ID != "" {
      continue
    }
    v.Bot = bot
    if n, ok := internalIDs[v.UserName]; ok {
      v.ID = strconv.FormatUint(initial+n, 10)
    } else if v.UserName == v.Bot.req.userName {
      v.ID = strconv.FormatUint(initial+idSelfOffset, 10)
    } else {
      // 生成一个ID并备注
      v.ID = strconv.FormatUint(ret.nextID(), 10)
      ret.Bot.req.Remark(v.UserName, v.ID)
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
    }
    ret.contacts.Store(v.UserName, v)
    ret.userNames.Store(v.ID, v.UserName)
  }
  // 第三次循环，处理群聊
  for _, v := range data {
    if v.Flag != ContactGroup {
      continue
    }
    v.Bot = bot
    // todo 群没有备注，默认用MaxID自增作为ID，然后用该ID和群名称建立对应关系来解决持久化问题，
    // todo 若群改名，会收到消息，需要在接收消息的地方处理
  }
  // 第四次循环，处理其他类型（ContactMPS等）的联系人，
  // 这类联系人没有ID，只能通过UserName/NickName或关键字索引，
  // 即只有UserName=>*Contact的对应关系，没有ID=>UserName的对应关系
  for _, v := range data {
    if v.ID == "" {
      v.Bot = bot
      ret.contacts.Store(v.UserName, v)
    }
  }
  if bot.Self != nil {
    bot.Self.ID = strconv.FormatUint(initial+idSelfOffset, 10)
  }
  return ret
}*/

func (cs *Contacts) Add(c *Contact) {
  if c == nil {
    return
  }
  cs.mu.Lock()
  defer cs.mu.Unlock()
  if v, ok := cs.userNameMap[c.UserName]; ok {
    delete(cs.userNameMap, v.UserName)
    delete(cs.idMap, v.Id)
    if c.Id == "" {
      c.Id = v.Id
    }
  }
  cs.userNameMap[c.UserName] = c
  if c.Id != "" {
    cs.idMap[c.Id] = c
  }
}

func (cs *Contacts) Remove(userName string) {
  if userName == "" {
    return
  }
  cs.mu.Lock()
  defer cs.mu.Unlock()
  if v, ok := cs.userNameMap[userName]; ok {
    delete(cs.userNameMap, userName)
    if v.Id != "" {
      delete(cs.idMap, v.Id)
    }
  }
}

func (cs *Contacts) Count() int {
  ret := 0
  cs.Each(func(*Contact) bool {
    ret++
    return true
  })
  return ret
}

func (cs *Contacts) FindById(id string) *Contact {
  if id == "" {
    return nil
  }
  cs.mu.RLock()
  defer cs.mu.RUnlock()
  if v, ok := cs.idMap[id]; ok {
    return v
  }
  return nil
}

func (cs *Contacts) FindByUserName(userName string) *Contact {
  if userName == "" {
    return nil
  }
  cs.mu.RLock()
  defer cs.mu.RUnlock()
  if v, ok := cs.userNameMap[userName]; ok {
    return v
  }
  return nil
}

func (cs *Contacts) FindByNickName(nickName string) *Contact {
  if nickName == "" {
    return nil
  }
  var ret *Contact
  cs.Each(func(c *Contact) bool {
    if nickName == c.NickName {
      ret = c
      return false
    }
    return true
  })
  return ret
}

// 根据Id/NickName关键字（3个字符及以上）模糊查找联系人
func (cs *Contacts) FindByKeyword(keyword string) []*Contact {
  if len(keyword) < 3 {
    return nil
  }
  ret := make([]*Contact, 0, 10)
  cs.Each(func(c *Contact) bool {
    if strings.Contains(c.Id, keyword) {
      ret = append(ret, c)
      return true
    }
    if strings.Contains(c.NickName, keyword) {
      ret = append(ret, c)
      return true
    }
    return true
  })
  if len(ret) == 0 {
    return nil
  }
  return ret
}

// 遍历所有联系人，action返回false表示终止遍历
func (cs *Contacts) Each(action func(*Contact) bool) {
  cs.mu.RLock()
  defer cs.mu.RUnlock()
  for _, v := range cs.userNameMap {
    if !action(v) {
      break
    }
  }
}

func (cs *Contacts) nextId() uint64 {
  cs.mu.Lock()
  cs.maxId++
  cs.mu.Unlock()
  return cs.maxId
}

func (cs *Contacts) initialId() uint64 {
  if id, ok := cs.Bot.Attr.Load(attrInitialId); ok {
    return id.(uint64)
  }
  str := strconv.FormatUint(cs.maxId, 10)
  str = str[:len(str)-4]
  ret, _ := strconv.ParseUint(str, 10, 64)
  ret *= 10000
  cs.Bot.Attr.Store(attrInitialId, ret)
  return ret
}

func getIdByRemarkName(remarkName string) uint64 {
  if remarkName == "" {
    return 0
  }
  ret, e := strconv.ParseUint(remarkName, 10, 64)
  if e != nil || ret <= idInitial+idGeneralOffset {
    return 0
  }
  return ret
}
