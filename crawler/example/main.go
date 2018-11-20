package main

import (
  "fmt"

  "github.com/kwf2030/commons/crawler"
)

func main() {
  e := crawler.GetRules().FromFiles([]string{"D:\\Workspace\\kwf2030\\commons\\crawler\\rule_test.yml"})
  if e != nil {
    panic(e)
  }
  e = crawler.LaunchChrome("")
  if e != nil {
    panic(e)
  }
  ch := crawler.Start()
  crawler.Enqueue([]*crawler.Page{{"1", "https://item.jd.com/2165601.html", "default", nil}})
  for pages := range ch {
    for _, p := range pages {
      fmt.Println("channel:", p.Id, p.Result)
    }
  }
}
