package crawler

import (
  "fmt"
  "strconv"
  "testing"
  "time"
)

func TestCrawler(t *testing.T) {
  e := GetRules().FromFiles([]string{"D:\\Workspace\\kwf2030\\commons\\crawler\\rule_test.yml"})
  if e != nil {
    t.Fatal(e)
  }
  e = LaunchChrome("")
  if e != nil {
    t.Fatal(e)
  }
  ch := Start()
  go func() {
    for pages := range ch {
      for _, p := range pages {
        fmt.Println(p.Id, p.Result)
      }
    }
  }()
  for i := 0; i < 5; i++ {
    Enqueue([]*Page{{strconv.Itoa(i), "https://item.jd.com/11684158.html", "default", nil}})
    time.Sleep(time.Second * 3)
  }
  Stop()
  time.Sleep(time.Second)
}
