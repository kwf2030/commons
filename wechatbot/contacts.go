package wechatbot

import (
  "sync"
)

type Contacts struct {
  data map[string]*Contact
  mu   *sync.RWMutex
  bot  *Bot
}

func initContacts(contacts []*Contact, bot *Bot) *Contacts {
  ret := &Contacts{
    data: make(map[string]*Contact, 5000),
    mu:   &sync.RWMutex{},
    bot:  bot,
  }
  for _, c := range contacts {
    c.withBot(bot)
    ret.data[c.UserName] = c
  }
  return ret
}

func (cs *Contacts) Add(c *Contact) {
  if c == nil || c.UserName == "" {
    return
  }
  cs.mu.Lock()
  cs.data[c.UserName] = c
  cs.mu.Unlock()
}

func (cs *Contacts) Remove(userName string) {
  if userName == "" {
    return
  }
  cs.mu.Lock()
  delete(cs.data, userName)
  cs.mu.Unlock()
}

func (cs *Contacts) Get(userName string) *Contact {
  if userName == "" {
    return nil
  }
  cs.mu.RLock()
  ret := cs.data[userName]
  cs.mu.RUnlock()
  return ret
}

func (cs *Contacts) Count() int {
  cs.mu.RLock()
  ret := len(cs.data)
  cs.mu.RUnlock()
  return ret
}

func (cs *Contacts) Each(f func(*Contact) bool) {
  cs.mu.RLock()
  defer cs.mu.RUnlock()
  for _, c := range cs.data {
    if !f(c) {
      break
    }
  }
}
