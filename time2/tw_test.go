package time2

import (
  "fmt"
  "sync"
  "testing"
  "time"
)

func TestTimingWheel(t *testing.T) {
  var wg sync.WaitGroup
  wg.Add(1)
  tw := NewTimingWheel(1000, time.Millisecond)
  tw.Start()
  tw.Delay(time.Second*5, nil, func(id uint64, _ interface{}) {
    fmt.Println(id, " delayed")
    wg.Done()
  })
  wg.Wait()
}
