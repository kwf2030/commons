package crawler

import (
  "fmt"
  "sync"
  "testing"
)

var wg sync.WaitGroup

var testRule = []byte(`id: "1"
version: 1
name: "jd"
alias: "京东"
priority: 100
group: "default"

patterns:
  - "jd.com"

page_load_timeout: "10s"

loop:
  name: "page"
  alias: "最新10页评论"
  export_cycle: 5
  prepare:
    eval: "{document.documentElement.scrollBy(0, 1000);Array.prototype.slice.call(document.querySelector('#detail > div > ul').children).filter(function (e) {return e.textContent.indexOf('商品评价') !== -1;})[0].click();true;}"
    wait_when_ready: "2s"
  eval: "JSON.stringify(Array.prototype.slice.call(document.querySelectorAll('.comment-con')).map(e=>e.textContent))"
  break: "count===10"
  next: "document.querySelector('.ui-pager-next').click()"
  wait: "2s"
`)

type H struct {
  name string
}

func (h *H) OnFields(p *Page, data map[string]string) {
  fmt.Println("======OnFields:", data)
}

func (h *H) OnLoop(p *Page, count int, data []string) {
  fmt.Println("======OnLoop:", count, data)
}

func (h *H) OnComplete(p *Page) {
  fmt.Println("======OnComplete")
  wg.Done()
}

func TestCrawler(t *testing.T) {
  e := Rules.FromBytes([][]byte{testRule})
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
