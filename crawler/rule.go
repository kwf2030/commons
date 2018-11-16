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
  errInvalidArgs = errors.New("invalid args")
  errNoRuleGroup = errors.New("no rule group")

  capacity = 16

  allRules = &Rules{groups: make(map[string][]*rule, capacity)}
)

type Rules struct {
  groups map[string][]*rule
  sync.RWMutex
}

func (r *Rules) match(group, url string) *rule {
  if group == "" || url == "" {
    return nil
  }
  r.RLock()
  defer r.RUnlock()
  rules, ok := r.groups[group]
  if !ok {
    return nil
  }
  for _, rule := range rules {
    for _, p := range rule.Patterns {
      if p.Regex.MatchString(url) {
        return rule
      }
    }
  }
  return nil
}

func (r *Rules) FromFiles(files []string) error {
  if len(files) == 0 {
    return errInvalidArgs
  }
  for _, f := range files {
    data, e := ioutil.ReadFile(f)
    if e != nil {
      return e
    }
    rr := &rule{}
    e = yaml.Unmarshal(data, r)
    if e != nil {
      return e
    }
    if rr.Group == "" {
      rr.Group = "default"
    }
    e = r.Update(rr.Group, []*rule{rr})
    if e != nil {
      return e
    }
  }
  return nil
}

func (r *Rules) Update(group string, rules []*rule) error {
  if group == "" || len(rules) == 0 {
    return errInvalidArgs
  }
  r.Lock()
  defer r.Unlock()
  if _, ok := r.groups[group]; !ok {
    r.groups[group] = make([]*rule, 0, capacity)
  }
  for _, rule := range rules {
    if rule.Group == "" {
      rule.Group = group
    }
    if rule.Group != group {
      continue
    }
    index := -1
    for i, old := range r.groups[group] {
      if old.Id == rule.Id {
        index = i
        break
      }
    }
    if index == -1 {
      r.groups[group] = append(r.groups[group], rule)
    } else {
      old := r.groups[group][index]
      if old.Version < rule.Version {
        r.groups[group][index] = rule
      }
    }
    initRule(rule)
  }
  sort.SliceStable(r.groups[group], func(i, j int) bool {
    return r.groups[group][i].Priority < r.groups[group][j].Priority
  })
  return nil
}

func (r *Rules) Remove(group string, ids ...string) error {
  if group == "" {
    return errInvalidArgs
  }
  r.Lock()
  defer r.Unlock()
  if _, ok := r.groups[group]; !ok {
    return errNoRuleGroup
  }
  if len(ids) == 0 {
    delete(r.groups, group)
    return nil
  }
  for _, id := range ids {
    index := -1
    for i, r := range r.groups[group] {
      if id == r.Id {
        index = i
        break
      }
    }
    if index != -1 {
      r.groups[group] = append(r.groups[group][:index], r.groups[group][index+1:]...)
    }
  }
  return nil
}

func initRule(rule *rule) {
  rule.Patterns = make([]*pattern, 0, len(rule.PatternsConf))
  for _, p := range rule.PatternsConf {
    re, e := regexp.Compile(p)
    if e != nil {
      continue
    }
    rule.Patterns = append(rule.Patterns, &pattern{
      Content: p,
      Regex:   re,
    })
  }
  rule.PageLoadTimeout = time.Second * 10
  if rule.PageLoadTimeoutConf != "" {
    rule.PageLoadTimeout, _ = time.ParseDuration(rule.PageLoadTimeoutConf)
  }
  if rule.Prepare != nil && rule.Prepare.WaitWhenReadyConf != "" {
    var e error
    rule.Prepare.WaitWhenReady, e = time.ParseDuration(rule.Prepare.WaitWhenReadyConf)
    if e != nil {
      rule.Prepare.WaitWhenReady = time.Second
    }
  }
  for _, v := range rule.Fields {
    if v.WaitConf != "" {
      var e error
      v.Wait, e = time.ParseDuration(v.WaitConf)
      if e != nil {
        v.Wait = time.Second
      }
    }
  }
  if rule.Loop != nil {
    var e error
    if rule.Loop.Prepare != nil && rule.Loop.Prepare.WaitWhenReadyConf != "" {
      rule.Loop.Prepare.WaitWhenReady, e = time.ParseDuration(rule.Loop.Prepare.WaitWhenReadyConf)
      if e != nil {
        rule.Loop.Prepare.WaitWhenReady = time.Second
      }
    }
    if rule.Loop.WaitConf != "" {
      rule.Loop.Wait, e = time.ParseDuration(rule.Loop.WaitConf)
      if e != nil {
        rule.Loop.Wait = time.Second
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
  Patterns            []*pattern    `yaml:"-"`
  PageLoadTimeoutConf string        `yaml:"page_load_timeout"`
  PageLoadTimeout     time.Duration `yaml:"-"`
  Prepare             *prepare      `yaml:"prepare"`
  Fields              []*field      `yaml:"fields"`
  Loop                *loop         `yaml:"loop"`
}

type pattern struct {
  Content string
  Regex   *regexp.Regexp
}

type prepare struct {
  Eval              string        `yaml:"eval"`
  WaitWhenReadyConf string        `yaml:"wait_when_ready"`
  WaitWhenReady     time.Duration `yaml:"-"`
}

type field struct {
  Name     string        `yaml:"name"`
  Alias    string        `yaml:"alias"`
  Value    string        `yaml:"value"`
  Eval     string        `yaml:"eval"`
  Export   bool          `yaml:"export"`
  WaitConf string        `yaml:"wait"`
  Wait     time.Duration `yaml:"-"`
}

type loop struct {
  Field       string        `yaml:"field"`
  Alias       string        `yaml:"alias"`
  ExportCycle int           `yaml:"export_cycle"`
  Prepare     *prepare      `yaml:"prepare"`
  Eval        string        `yaml:"eval"`
  Next        string        `yaml:"next"`
  WaitConf    string        `yaml:"wait"`
  Wait        time.Duration `yaml:"-"`
  Break       string        `yaml:"break"`
}
