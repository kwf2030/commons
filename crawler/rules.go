package crawler

import (
  "errors"
  "io/ioutil"
  "regexp"
  "sort"
  "sync"
  "time"

  "gopkg.in/yaml.v2"
)

var (
  ErrArgsInvalid   = errors.New("args invalid")
  ErrGroupNotFound = errors.New("group not found")

  rules = &Rules{groups: make(map[string][]*rule, 16)}
)

type Rules struct {
  groups map[string][]*rule
  sync.RWMutex
}

func (rs *Rules) match(group, url string) *rule {
  if group == "" || url == "" {
    return nil
  }
  rs.RLock()
  defer rs.RUnlock()
  arr, ok := rs.groups[group]
  if !ok {
    return nil
  }
  for _, r := range arr {
    for _, p := range r.patterns {
      if p.regex.MatchString(url) {
        return r
      }
    }
  }
  return nil
}

func (rs *Rules) FromFiles(files []string) error {
  if len(files) == 0 {
    return ErrArgsInvalid
  }
  for _, f := range files {
    data, e := ioutil.ReadFile(f)
    if e != nil {
      return e
    }
    r := &rule{Group: "default"}
    e = yaml.Unmarshal(data, r)
    if e != nil {
      return e
    }
    initRule(r)
    e = rs.update(r.Group, []*rule{r})
    if e != nil {
      return e
    }
  }
  return nil
}

func (rs *Rules) update(group string, arr []*rule) error {
  if group == "" || len(arr) == 0 {
    return ErrArgsInvalid
  }
  rs.Lock()
  defer rs.Unlock()
  if _, ok := rs.groups[group]; !ok {
    rs.groups[group] = make([]*rule, 0, 16)
  }
  for _, r := range arr {
    if r.Group == "" {
      r.Group = group
    }
    if r.Group != group {
      continue
    }
    index := -1
    for i, old := range rs.groups[group] {
      if old.Id == r.Id {
        index = i
        break
      }
    }
    if index == -1 {
      rs.groups[group] = append(rs.groups[group], r)
    } else {
      old := rs.groups[group][index]
      if old.Version < r.Version {
        rs.groups[group][index] = r
      }
    }
  }
  sort.SliceStable(rs.groups[group], func(i, j int) bool {
    return rs.groups[group][i].Priority < rs.groups[group][j].Priority
  })
  return nil
}

func (rs *Rules) Remove(group string, ids ...string) error {
  if group == "" {
    return ErrArgsInvalid
  }
  rs.Lock()
  defer rs.Unlock()
  if _, ok := rs.groups[group]; !ok {
    return ErrGroupNotFound
  }
  if len(ids) == 0 {
    delete(rs.groups, group)
    return nil
  }
  for _, id := range ids {
    index := -1
    for i, r := range rs.groups[group] {
      if id == r.Id {
        index = i
        break
      }
    }
    if index != -1 {
      rs.groups[group] = append(rs.groups[group][:index], rs.groups[group][index+1:]...)
    }
  }
  return nil
}

func initRule(rule *rule) {
  rule.patterns = make([]*pattern, 0, len(rule.PatternsConf))
  for _, p := range rule.PatternsConf {
    re, e := regexp.Compile(p)
    if e != nil {
      continue
    }
    rule.patterns = append(rule.patterns, &pattern{p, re})
  }
  rule.pageLoadTimeout = time.Second * 10
  if rule.PageLoadTimeoutConf != "" {
    rule.pageLoadTimeout, _ = time.ParseDuration(rule.PageLoadTimeoutConf)
  }
  if rule.Prepare != nil && rule.Prepare.WaitWhenReadyConf != "" {
    var e error
    rule.Prepare.waitWhenReady, e = time.ParseDuration(rule.Prepare.WaitWhenReadyConf)
    if e != nil {
      rule.Prepare.waitWhenReady = time.Second
    }
  }
  for _, f := range rule.Fields {
    if f.WaitConf != "" {
      var e error
      f.wait, e = time.ParseDuration(f.WaitConf)
      if e != nil {
        f.wait = time.Second
      }
    }
  }
  if rule.Loop != nil {
    var e error
    if rule.Loop.Prepare != nil && rule.Loop.Prepare.WaitWhenReadyConf != "" {
      rule.Loop.Prepare.waitWhenReady, e = time.ParseDuration(rule.Loop.Prepare.WaitWhenReadyConf)
      if e != nil {
        rule.Loop.Prepare.waitWhenReady = time.Second
      }
    }
    if rule.Loop.WaitConf != "" {
      rule.Loop.wait, e = time.ParseDuration(rule.Loop.WaitConf)
      if e != nil {
        rule.Loop.wait = time.Second
      }
    }
  }
}

type rule struct {
  Id                  string        `yaml:"id"`
  Version             int           `yaml:"version"`
  Name                string        `yaml:"name"`
  Alias               string        `yaml:"alias"`
  Group               string        `yaml:"group"`
  Priority            int           `yaml:"priority"`
  PatternsConf        []string      `yaml:"patterns"`
  patterns            []*pattern    `yaml:"-"`
  PageLoadTimeoutConf string        `yaml:"page_load_timeout"`
  pageLoadTimeout     time.Duration `yaml:"-"`
  Prepare             *prepare      `yaml:"prepare"`
  Fields              []*field      `yaml:"fields"`
  Loop                *loop         `yaml:"loop"`
}

type pattern struct {
  content string
  regex   *regexp.Regexp
}

type prepare struct {
  Eval              string        `yaml:"eval"`
  WaitWhenReadyConf string        `yaml:"wait_when_ready"`
  waitWhenReady     time.Duration `yaml:"-"`
}

type field struct {
  Name     string        `yaml:"name"`
  Alias    string        `yaml:"alias"`
  Value    string        `yaml:"value"`
  Eval     string        `yaml:"eval"`
  Export   bool          `yaml:"export"`
  WaitConf string        `yaml:"wait"`
  wait     time.Duration `yaml:"-"`
}

type loop struct {
  Name        string        `yaml:"name"`
  Alias       string        `yaml:"alias"`
  ExportCycle int           `yaml:"export_cycle"`
  Prepare     *prepare      `yaml:"prepare"`
  Eval        string        `yaml:"eval"`
  Next        string        `yaml:"next"`
  WaitConf    string        `yaml:"wait"`
  wait        time.Duration `yaml:"-"`
  Break       string        `yaml:"break"`
}
