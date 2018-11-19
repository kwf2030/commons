package crawler

import (
  "fmt"
  "strconv"
  "testing"
  "time"
)

func TestCrawler(t *testing.T) {
  e := GetRules().FromFiles([]string{"D:\\Workspace\\kwf2030\\commons\\crawler\\jd.yml"})
  if e != nil {
    t.Fatal(e)
  }
  e = LaunchChrome("")
  if e != nil {
    t.Fatal(e)
  }
  ch := Start()
  go func() {
    c := time.Tick(time.Second * 10)
    id := 0
    for range c {
      id++
      Enqueue([]*Page{{strconv.Itoa(id), "https://item.jd.com/11684158.html", nil}})
    }
  }()
  for pages := range ch {
    for _, p := range pages {
      fmt.Println(p.Id, p.Result)
    }
  }
}
