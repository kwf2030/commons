package main

import (
  "encoding/json"
  "errors"
  "io/ioutil"
  "os"
  "runtime"
)

func main() {
  // debugResTable()
  // debugResTable2()
  // debugXml()
  debugXml2()
}

func debugResTable() {
  var file string
  switch runtime.GOOS {
  case "windows":
    file = "C:\\Users\\WangFeng\\Desktop\\resources.arsc"
  case "linux":
    file = "/home/wangfeng/workspace/wechat/raw/resources.arsc"
  default:
    panic(errors.New("os not supported"))
  }
  rt := ParseResTable(file)
  if rt != nil {
    data, e := json.Marshal(rt)
    if e != nil {
      panic(e)
    }
    e = ioutil.WriteFile("C:\\Users\\WangFeng\\Desktop\\table.json", data, os.ModePerm)
    if e != nil {
      panic(e)
    }
  }
}

func debugResTable2() {
  var file string
  switch runtime.GOOS {
  case "windows":
    file = "C:\\Users\\WangFeng\\Desktop\\resources.arsc"
  case "linux":
    file = "/home/wangfeng/workspace/wechat/raw/resources.arsc"
  default:
    panic(errors.New("os not supported"))
  }
  rt := ParseResTable(file)
  if rt != nil {
    rt2 := NewResTable2(rt)
    data, e := json.Marshal(rt2)
    if e != nil {
      panic(e)
    }
    e = ioutil.WriteFile("C:\\Users\\WangFeng\\Desktop\\table2.json", data, os.ModePerm)
    if e != nil {
      panic(e)
    }
  }
}

func debugXml() {
  var file string
  switch runtime.GOOS {
  case "windows":
    file = "C:\\Users\\WangFeng\\Desktop\\AndroidManifest.xml"
  case "linux":
    file = "/home/wangfeng/workspace/wechat/raw/AndroidManifest.xml"
  default:
    panic(errors.New("os not supported"))
  }
  xml := ParseXml(file)
  if xml != nil {
    data, e := json.Marshal(xml)
    if e != nil {
      panic(e)
    }
    e = ioutil.WriteFile("C:\\Users\\WangFeng\\Desktop\\xml.json", data, os.ModePerm)
    if e != nil {
      panic(e)
    }
  }
}

func debugXml2() {
  var file string
  switch runtime.GOOS {
  case "windows":
    file = "C:\\Users\\WangFeng\\Desktop\\AndroidManifest.xml"
  case "linux":
    file = "/home/wangfeng/workspace/wechat/raw/AndroidManifest.xml"
  default:
    panic(errors.New("os not supported"))
  }
  xml := ParseXml(file)
  if xml != nil {
    xml2 := NewXml2(xml)
    data, e := json.Marshal(xml2)
    if e != nil {
      panic(e)
    }
    e = ioutil.WriteFile("C:\\Users\\WangFeng\\Desktop\\xml2.json", data, os.ModePerm)
    if e != nil {
      panic(e)
    }
  }
}
