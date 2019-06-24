package time2

import (
  "fmt"
  "sync"
  "testing"
  "time"

  "github.com/kwf2030/commons/base"
)

func TestNormal(t *testing.T) {
  var wg sync.WaitGroup
  wg.Add(3)
  DefaultTimingWheel.Start()
  DefaultTimingWheel.Delay(time.Second*10, "data1", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  time.Sleep(time.Second * 5)
  DefaultTimingWheel.Delay(time.Second*8, "data2", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  DefaultTimingWheel.Delay(time.Millisecond*200, "data3", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  wg.Wait()
}

func TestCritical1(t *testing.T) {
  var wg sync.WaitGroup
  wg.Add(3)
  DefaultTimingWheel.Start()
  DefaultTimingWheel.Delay(time.Second*59, "data1", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  time.Sleep(time.Second * 5)
  DefaultTimingWheel.Delay(time.Second*59, "data2", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  DefaultTimingWheel.Delay(time.Millisecond*100, "data3", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  wg.Wait()
}

func TestCritical2(t *testing.T) {
  var wg sync.WaitGroup
  wg.Add(3)
  DefaultTimingWheel.Start()
  DefaultTimingWheel.Delay(time.Second*60, "data1", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  time.Sleep(time.Millisecond * 8500)
  DefaultTimingWheel.Delay(time.Second*60, "data2", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  DefaultTimingWheel.Delay(time.Millisecond*500, "data3", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  wg.Wait()
}

func TestCritical3(t *testing.T) {
  var wg sync.WaitGroup
  wg.Add(3)
  DefaultTimingWheel.Start()
  DefaultTimingWheel.Delay(time.Second*61, "data1", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  time.Sleep(time.Millisecond * 3765)
  DefaultTimingWheel.Delay(time.Second*61, "data2", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  DefaultTimingWheel.Delay(time.Millisecond*1000, "data3", func(id uint64, data interface{}) {
    fmt.Printf("executing task %d: %v\n", id, data)
    wg.Done()
  })
  wg.Wait()
}

func TestConcurrent(t *testing.T) {
  n := 10000
  var wg sync.WaitGroup
  wg.Add(n)
  DefaultTimingWheel.Start()
  for i := 0; i < n; i++ {
    t := time.Second * time.Duration(base.R.Intn(100)+1)
    DefaultTimingWheel.Delay(t, i, func(id uint64, data interface{}) {
      fmt.Printf("executing task %d: %v\n", id, data)
      wg.Done()
    })
  }
  wg.Wait()
}

func TestCustom(t *testing.T) {
  tw := NewTimingWheel(100, time.Millisecond*100)
  n := 1000
  var wg sync.WaitGroup
  wg.Add(n)
  tw.Start()
  for i := 0; i < n; i++ {
    t := time.Millisecond * time.Duration(base.R.Intn(1000)+1)
    tw.Delay(t, i, func(id uint64, data interface{}) {
      fmt.Printf("executing task %d: %v\n", id, data)
      wg.Done()
    })
  }
  wg.Wait()
}
