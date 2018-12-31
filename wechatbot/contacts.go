package wechatbot

import (
  "github.com/kwf2030/commons/times"
  "strconv"
  "strings"
  "sync"
  "time"
)

const (
  idInitial      = uint64(2018E4)
  idOffset       = uint64(1E7)
  idOffsetFriend = uint64(300)
  idOffsetSelf   = uint64(1E6)
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

func initContacts(contacts []*Contact, bot *Bot) *Contacts {
  ret := &Contacts{
    mu:          &sync.RWMutex{},
    userNameMap: make(map[string]*Contact, 5000),
    idMap:       make(map[string]*Contact, 5000),
    Bot:         bot,
  }
  if b, ok := bot.Attr.Load(attrIdEnabled); !ok || !b.(bool) {
    for _, c := range contacts {
      c.withBot(bot)
    }
    return ret
  }
  // 第1次循环，处理已备注的好友
  for _, c := range contacts {
    c.withBot(bot)
    if c.Type != ContactFriend {
      continue
    }
    if id := getIdByRemarkName(c.RemarkName); id != 0 {
      if ret.maxId < id {
        ret.maxId = id
      }
      c.Id = strconv.FormatUint(id, 10)
      ret.userNameMap[c.UserName] = c
      ret.idMap[c.Id] = c
    }
  }
  // 如果Bot从未设置过联系人的备注，那么起始Id是根据当前已经运行的Bot的个数来决定的
  // 这是为了当有多个Bot的时候，每个Bot的联系人Id唯一且不重复
  if ret.maxId == 0 {
    ret.maxId = idInitial + (uint64(CountBots()-1) * idOffset) + idOffsetFriend
  }
  // 第2次循环，处理未备注的好友或系统账号
  initialId := ret.initialId()
  for _, c := range contacts {
    if c.Id != "" || (c.Type != ContactFriend && c.Type != ContactSystem) {
      continue
    }
    if id, ok := internalIds[c.UserName]; ok {
      c.Id = strconv.FormatUint(initialId+id, 10)
    } else if c.UserName == c.Bot.req.UserName {
      c.Id = strconv.FormatUint(initialId+idOffsetSelf, 10)
    } else {
      // 生成一个Id并备注
      c.Id = strconv.FormatUint(ret.nextId(), 10)
      ret.Bot.req.Remark(c.UserName, c.Id)
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
    }
    ret.userNameMap[c.UserName] = c
    ret.idMap[c.Id] = c
  }
  // 第3次循环，处理群聊
  for _, c := range contacts {
    if c.Type != ContactGroup {
      continue
    }
    // todo 群没有备注，默认用MaxID自增作为ID，然后用该ID和群名称建立对应关系来解决持久化问题，
    // todo 若群改名，会收到消息，需要在接收消息的地方处理
  }
  // 第4次循环，处理其他类型的联系人，
  // 这类联系人没有Id，只能通过UserName/NickName或关键字查找，
  // 即只有UserName=>*Contact的对应关系，没有Id=>*Contact的对应关系
  for _, c := range contacts {
    if c.Id == "" {
      ret.userNameMap[c.UserName] = c
    }
  }
  return ret
}

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

func (cs *Contacts) nextId() uint64 {
  cs.mu.Lock()
  cs.maxId++
  cs.mu.Unlock()
  return cs.maxId
}

func getIdByRemarkName(remarkName string) uint64 {
  if remarkName == "" {
    return 0
  }
  ret, e := strconv.ParseUint(remarkName, 10, 64)
  if e != nil || ret <= idInitial+idOffsetFriend {
    return 0
  }
  return ret
}
