package main

import (
  "errors"
  "fmt"
  "runtime"
  "strconv"
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
    file = "/home/wangfeng/workspace/wechat/raw/resources.arsc"
  default:
    panic(errors.New("os not supported"))
  }
  ParseResTable(file)
  /*table := ParseResTable(file)
  printTableInfo(table)*/
}

func printTableInfo(table *ResTable) {
  fmt.Println("==========Table==========")
  fmt.Println("Type:", table.Type)
  fmt.Println("Size:", table.Size)
  fmt.Println("HeaderSize:", table.HeaderSize)
  fmt.Println("==========String Pool==========")
  fmt.Println("Type:", table.StrPool.Type)
  fmt.Println("Size:", table.StrPool.Size)
  fmt.Println("HeaderSize:", table.StrPool.HeaderSize)
  fmt.Println("StrCount:", table.StrPool.StrCount)
  fmt.Println("StrStart:", table.StrPool.StrStart)
  fmt.Println("StyleCount:", table.StrPool.StyleCount)
  fmt.Println("StyleStart:", table.StrPool.StyleStart)
  fmt.Println("==========Package==========")
  fmt.Println("Type:", table.Packages[0].Type)
  fmt.Println("Size:", table.Packages[0].Size)
  fmt.Println("HeaderSize:", table.Packages[0].HeaderSize)
  fmt.Println("Types:", table.Packages[0].TypeStrPool.Strs)
  fmt.Printf("TypeCount/TypeStrCount: %d/%d\n", table.Packages[0].TypeCount, table.Packages[0].TypeStrPool.StrCount)
  fmt.Println("TypeStrPoolStart:", table.Packages[0].TypeStrPoolStart)
  fmt.Printf("KeyCount/KeyStrCount: %d/%d\n", table.Packages[0].KeyCount, table.Packages[0].KeyStrPool.StrCount)
  fmt.Println("KeyStrPoolStart:", table.Packages[0].KeyStrPoolStart)
  for i, v := range table.Packages[0].Types {
    fmt.Println("==========Types[" + strconv.Itoa(i) + "]==========")
    fmt.Println("Type:", v.Type)
    fmt.Println("Size:", v.Size)
    fmt.Println("HeaderSize:", v.HeaderSize)
    fmt.Println("Id:", v.Id)
    fmt.Println("EntryCount:", v.EntryCount)
    fmt.Println("EntryStart:", v.EntryStart)
    fmt.Println("ConfigSize:", v.Config.Size)
  }
  for i, v := range table.Packages[0].TypeSpecs {
    fmt.Println("==========Type Spec[" + strconv.Itoa(i) + "]==========")
    fmt.Println("Type:", v.Type)
    fmt.Println("Size:", v.Size)
    fmt.Println("HeaderSize:", v.HeaderSize)
    fmt.Println("Id:", v.Id)
    fmt.Println("EntryCount:", v.EntryCount)
  }
}
