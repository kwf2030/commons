package crawler

import (
  "encoding/json"
  "fmt"
  "sync"
  "testing"
)

var wg sync.WaitGroup

type H struct {
  name string
}

func (h *H) OnFields(p *Page, data map[string]string) {
  fmt.Println("======OnFields:", data)
}

func (h *H) OnLoop(p *Page, count int, data []string) {
  fmt.Println("======OnLoop:", count, data)
  for _, v := range data {
    arr := make([]string, 0, 10)
    e := json.Unmarshal([]byte(v), &arr)
    if e != nil {
      panic(e)
    }
    fmt.Println("len=", len(arr))
  }
}

func (h *H) OnComplete(p *Page) {
  fmt.Println("======OnComplete")
  wg.Done()
}

func TestCrawler(t *testing.T) {
  e := Rules.FromFiles([]string{"D:\\Workspace\\kwf2030\\commons\\crawler\\rule_test.yml"})
  if e != nil {
    t.Fatal(e)
  }
  e = LaunchChrome("")
  if e != nil {
    t.Fatal(e)
  }
  wg.Add(1)
  h := &H{"JingDong"}
  p := NewPage("1", "https://item.jd.com/6655821.html", "default")
  e = p.Crawl(h)
  if e != nil {
    panic(e)
  }
  wg.Wait()
}
