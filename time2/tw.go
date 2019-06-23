package time2

import (
  "math"
  "sync"
  "time"

  "github.com/kwf2030/commons/base"
)

const (
  stateReady = iota
  stateRunning
  stateStopped
)

var DefaultTimingWheel = NewTimingWheel(60, time.Second)

type TimingWheel struct {
  // 单个slot时间
  duration time.Duration
  ticker   *time.Ticker

  // 当前所在slot
  cur uint64

  // 每个slot对应一个bucket，即一轮共有len(buckets)个slot，
  // 每个bucket是一个map，包含该slot的所有task
  buckets []map[uint64]*task
  // 一轮有多少个slot，等于len(buckets)
  slots uint64

  // 计时器停止信号
  stopChan chan struct{}

  state int

  mu sync.Mutex
}

func NewTimingWheel(slots int, duration time.Duration) *TimingWheel {
  if slots <= 0 {
    return nil
  }
  buckets := make([]map[uint64]*task, slots)
  for i := range buckets {
    buckets[i] = make(map[uint64]*task, 16)
  }
  return &TimingWheel{
    duration: duration,
    ticker:   time.NewTicker(duration),
    cur:      0,
    buckets:  buckets,
    slots:    uint64(slots),
    stopChan: make(chan struct{}),
    state:    stateReady,
    mu:       sync.Mutex{},
  }
}

func (tw *TimingWheel) Start() {
  tw.mu.Lock()
  defer tw.mu.Unlock()
  if tw.state == stateReady {
    tw.state = stateRunning
    go tw.run()
  }
}

func (tw *TimingWheel) Stop() {
  tw.mu.Lock()
  defer tw.mu.Unlock()
  if tw.state == stateRunning {
    tw.state = stateStopped
    close(tw.stopChan)
  }
}

func (tw *TimingWheel) Delay(delay time.Duration, data interface{}, f func(uint64, interface{})) uint64 {
  n1 := int64(delay/tw.duration) / int64(tw.slots)
  n2 := uint64(delay/tw.duration) % tw.slots
  if n2 == 0 {
    n2 = 1
  }
  tw.mu.Lock()
  defer tw.mu.Unlock()
  n := tw.cur + n2
  task := &task{
    id:    (n << 32) | (base.R.Uint64() >> 32),
    round: n1,
    data:  data,
    f:     f,
  }
  tw.buckets[n][task.id] = task
  return task.id
}

func (tw *TimingWheel) At(t time.Time, data interface{}, f func(uint64, interface{})) uint64 {
  now := UTC()
  if t.Before(now) {
    return 0
  }
  return tw.Delay(t.Sub(now), data, f)
}

func (tw *TimingWheel) Cancel(id uint64) {
  i := id >> 32
  if i < tw.slots {
    tw.mu.Lock()
    defer tw.mu.Unlock()
    delete(tw.buckets[i], id)
  }
}

func (tw *TimingWheel) run() {
  for {
    select {
    case <-tw.ticker.C:
      tw.mu.Lock()
      if tw.cur == tw.slots-1 {
        tw.cur = 0
      } else {
        tw.cur++
      }
      tw.tick()
      tw.mu.Unlock()

    case <-tw.stopChan:
      tw.ticker.Stop()
      return
    }
  }
}

func (tw *TimingWheel) tick() {
  tasks := make([]*task, 0, 2)
  for _, t := range tw.buckets[tw.cur] {
    if t.round <= 0 {
      tasks = append(tasks, t)
    }
    if t.round > math.MinInt64 {
      t.round--
    }
  }
  for _, t := range tasks {
    delete(tw.buckets[tw.cur], t.id)
    go t.f(t.id, t.data)
  }
}

type task struct {
  // 高32位表示task所在的slot，低32位随机生成
  id uint64

  // 剩余轮数
  round int64

  data interface{}

  f func(uint64, interface{})
}
