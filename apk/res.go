package main

import (
  "errors"
  "fmt"
  "runtime"
)

type ResStrStyle struct {
  Ref   ResStrRef
  Spans []ResStrSpan
}

type ResStrRef struct {
  Index uint32
}

type ResStrSpan struct {
  // 样式字符串在字符串池中的偏移
  Name ResStrRef

  // 应用样式的第一个字符
  FirstChar uint32

  // 应用样式的最后一个字符
  LastChar uint32
}

func main() {
  var file string
  switch runtime.GOOS {
  case "windows":
    file = "C:\\Users\\WangFeng\\Desktop\\resources.arsc"
  case "linux":
    file = "/home/wangfeng/workspace/wechat/tmp/resources.arsc"
  default:
    panic(errors.New("os not supported"))
  }
  table := ParseResTable(file)
  printTableInfo(table)
}

func printTableInfo(table *ResTable) {
  fmt.Println("==========Table==========")
  fmt.Println("Type:", table.Header.Type)
  fmt.Println("Size:", table.Header.Size)
  fmt.Println("HeaderSize:", table.Header.HeaderSize)
  fmt.Println("==========String Pool==========")
  fmt.Println("Type:", table.StrPool.Type)
  fmt.Println("Size:", table.StrPool.Size)
  fmt.Println("HeaderSize:", table.StrPool.HeaderSize)
  fmt.Println("StrCount:", table.StrPool.StrCount)
  fmt.Println("StrStart:", table.StrPool.StrStart)
  fmt.Println("StyleCount:", table.StrPool.StyleCount)
  fmt.Println("StyleStart:", table.StrPool.StyleStart)
  fmt.Println("==========Package==========")
  fmt.Println("Type:", table.Package.Type)
  fmt.Println("Size:", table.Package.Size)
  fmt.Println("HeaderSize:", table.Package.HeaderSize)
  fmt.Printf("TypeCount/TypeStrCount: %d/%d\n", table.Package.TypeCount, table.Package.TypeStrPool.StrCount)
  fmt.Println("TypeStrPoolStart:", table.Package.TypeStrPoolStart)
  fmt.Printf("EntryCount/EntryStrCount: %d/%d\n", table.Package.EntryCount, table.Package.EntryStrPool.StrCount)
  fmt.Println("EntryStrPoolStart:", table.Package.EntryStrPoolStart)
}
