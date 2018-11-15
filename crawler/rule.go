package crawler

import (
  "errors"
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

  allRules = &Rules{groups: make(map[string][]*Rule, capacity)}
)

type Rules struct {
  groups map[string][]*Rule
  sync.RWMutex
}

func (r *Rules) Match(group, url string) *Rule {
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

func (r *Rules) Update(group string, data []byte) error {
  if group == "" || len(data) == 0 {
    return errInvalidArgs
  }
  rules := make([]*Rule, 0, capacity)
  e := yaml.Unmarshal(data, &rules)
  if e != nil {
    return e
  }
  r.Lock()
  defer r.Unlock()
  if _, ok := r.groups[group]; !ok {
    r.groups[group] = make([]*Rule, 0, capacity)
  }
  for _, rule := range rules {
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

func initRule(rule *Rule) {
  rule.Patterns = make([]*Pattern, 0, len(rule.PatternsConf))
  for _, pattern := range rule.PatternsConf {
    re, e := regexp.Compile(pattern)
    if e != nil {
      continue
    }
    rule.Patterns = append(rule.Patterns, &Pattern{
      Content: pattern,
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

type Rule struct {
  Id                  string        `yaml:"id"`
  Version             int           `yaml:"version"`
  Name                string        `yaml:"name"`
  Alias               string        `yaml:"alias"`
  Group               string        `yaml:"group"`
  Priority            int           `yaml:"priority"`
  PatternsConf        []string      `yaml:"patterns"`
  Patterns            []*Pattern    `yaml:"-"`
  PageLoadTimeoutConf string        `yaml:"page_load_timeout"`
  PageLoadTimeout     time.Duration `yaml:"-"`
  Prepare             *Prepare      `yaml:"prepare"`
  Fields              []*Field      `yaml:"fields"`
  Loop                *Loop         `yaml:"loop"`
}

type Pattern struct {
  Content string
  Regex   *regexp.Regexp
}

type Prepare struct {
  Eval              string        `yaml:"eval"`
  WaitWhenReadyConf string        `yaml:"wait_when_ready"`
  WaitWhenReady     time.Duration `yaml:"-"`
}

type Field struct {
  Name     string        `yaml:"name"`
  Alias    string        `yaml:"alias"`
  Value    string        `yaml:"value"`
  Eval     string        `yaml:"eval"`
  Export   bool          `yaml:"export"`
  WaitConf string        `yaml:"wait"`
  Wait     time.Duration `yaml:"-"`
}

type Loop struct {
  Field       string        `yaml:"field"`
  Alias       string        `yaml:"alias"`
  ExportCycle int           `yaml:"export_cycle"`
  Prepare     *Prepare      `yaml:"prepare"`
  Eval        string        `yaml:"eval"`
  Next        string        `yaml:"next"`
  WaitConf    string        `yaml:"wait"`
  Wait        time.Duration `yaml:"-"`
  Break       string        `yaml:"break"`
}
