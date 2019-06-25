package time2

import (
  "sync"
  "time"

  "github.com/kwf2030/commons/base"
)

const (
  ready = iota
  running
  stopped
)

var DefaultTimingWheel = NewTimingWheel(60, time.Second)

type TimingWheel struct {
  state int

  ticker *time.Ticker

  // 单个slot时间（duration per slot）
  dps time.Duration

  // 当前slot
  slot uint64
  // 一轮有多少个slot
  slots uint64

  // 每个slot对应一个bucket，一轮共有len(buckets)个slot，
  // 每个bucket是一个map，包含该slot的所有task
  buckets []map[uint64]*task

  // 计时器停止信号
  stopChan chan struct{}

  mu sync.Mutex
}

func NewTimingWheel(slots int, dps time.Duration) *TimingWheel {
  if slots <= 0 || dps <= 0 {
    return nil
  }
  buckets := make([]map[uint64]*task, slots)
  for i := range buckets {
    buckets[i] = make(map[uint64]*task, 16)
  }
  return &TimingWheel{
    state:    ready,
    ticker:   time.NewTicker(dps),
    dps:      dps,
    slot:     0,
    slots:    uint64(slots),
    buckets:  buckets,
    stopChan: make(chan struct{}),
    mu:       sync.Mutex{},
  }
}

func (tw *TimingWheel) Start() {
  tw.mu.Lock()
  defer tw.mu.Unlock()
  if tw.state == ready {
    tw.state = running
    go tw.run()
  }
}

func (tw *TimingWheel) Stop() {
  tw.mu.Lock()
  defer tw.mu.Unlock()
  if tw.state == running {
    tw.state = stopped
    close(tw.stopChan)
  }
}

func (tw *TimingWheel) Delay(delay time.Duration, data interface{}, f func(uint64, interface{})) uint64 {
  if delay <= 0 || f == nil {
    return 0
  }
  // 剩余轮数
  n1 := uint64(delay/tw.dps) / tw.slots
  // 需要几个slot的时间
  n2 := uint64(delay/tw.dps) % tw.slots
  if n2 == 0 {
    // 不足一个slot时间，按一个slot时间算
    n2 = 1
  }
  tw.mu.Lock()
  defer tw.mu.Unlock()
  n := (tw.slot + n2 + 1) % tw.slots
  task := &task{
    id:    (n << 32) | (base.Rand.Uint64() >> 32),
    round: n1,
    data:  data,
    f:     f,
  }
  tw.buckets[n][task.id] = task
  return task.id
}

func (tw *TimingWheel) At(t time.Time, data interface{}, f func(uint64, interface{})) uint64 {
  if f == nil {
    return 0
  }
  if now := UTC(); t.After(now) {
    return tw.Delay(t.Sub(now), data, f)
  }
  return 0
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
    case <-tw.stopChan:
      tw.ticker.Stop()
      return

    case <-tw.ticker.C:
      tw.mu.Lock()
      if tw.slot == tw.slots-1 {
        tw.slot = 0
      } else {
        tw.slot++
      }
      tw.tick()
      tw.mu.Unlock()
    }
  }
}

func (tw *TimingWheel) tick() {
  tasks := make([]*task, 0, 16)
  for _, t := range tw.buckets[tw.slot] {
    if t.round > 0 {
      t.round--
    } else {
      tasks = append(tasks, t)
    }
  }
  for _, t := range tasks {
    delete(tw.buckets[tw.slot], t.id)
    go t.f(t.id, t.data)
  }
}

type task struct {
  // 高32位表示task所在的slot，低32位随机生成
  id uint64

  // 剩余轮数
  round uint64

  data interface{}

  f func(uint64, interface{})
}
