package crawler

import (
  "encoding/json"
  "fmt"
  "strconv"
  "testing"
)

func TestCrawler(t *testing.T) {
  e := Rules.FromFiles([]string{"D:\\Workspace\\kwf2030\\commons\\crawler\\rule_test.yml"})
  if e != nil {
    t.Fatal(e)
  }
  e = LaunchChrome("")
  if e != nil {
    t.Fatal(e)
  }
  for i := 0; i < 2; i++ {
    Crawl(&Page{strconv.Itoa(i), "https://item.jd.com/2165601.html", "default"},
      func(page *Page, result map[string]interface{}) {
        fmt.Println("field", page.Id, result)
      },
      func(page *Page, count int, result []string) {
        fmt.Println("loop", page.Id, count, result)
      })
  }
}

func TestCrawler2(t *testing.T) {
  e := Rules.FromFiles([]string{"D:\\Workspace\\kwf2030\\commons\\crawler\\rule_test.yml"})
  if e != nil {
    t.Fatal(e)
  }
  e = LaunchChrome("")
  if e != nil {
    t.Fatal(e)
  }
  Crawl(&Page{"100", "https://www.amazon.cn/gp/registry/wishlist/ref=nav_wishlist_btn", "default"},
    func(page *Page, result map[string]interface{}) {
      obj := result["products"].(string)
      m := make([]map[string]interface{}, 200)
      e := json.Unmarshal([]byte(obj), &m)
      if e != nil {
        panic(e)
      }
      fmt.Println(len(m))
    },
    func(page *Page, count int, result []string) {
      fmt.Println("loop", page.Id, count, result)
    })
}
