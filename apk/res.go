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
  arsc := ParseResArsc(file)
  for i := 0; i < 10; i++ {
    fmt.Println(arsc.GlobalStrPoolChunk.Strs[i])
  }
}
