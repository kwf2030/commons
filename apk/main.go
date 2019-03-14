package main

import (
  "encoding/json"
  "errors"
  "io/ioutil"
  "os"
  "runtime"
)

func main() {
  //debugResTable()
  debugManifest()
}

func debugResTable() {
  var inFile, outFile string
  switch runtime.GOOS {
  case "windows":
    inFile = "C:\\Users\\WangFeng\\Desktop\\resources.arsc"
    outFile = "C:\\Users\\WangFeng\\Desktop\\resources.json"
  case "linux":
    inFile = "/home/wangfeng/workspace/wechat/raw/resources.arsc"
    outFile = "/home/wangfeng/workspace/wechat/raw/resources.json"
  default:
    panic(errors.New("os not supported"))
  }
  rt := ParseResTable(inFile)
  if rt != nil {
    rt2 := NewResTable2(rt)
    data, e := json.Marshal(rt2)
    if e != nil {
      panic(e)
    }
    e = ioutil.WriteFile(outFile, data, os.ModePerm)
    if e != nil {
      panic(e)
    }
  }
}

func debugManifest() {
  var inFile, outFile string
  switch runtime.GOOS {
  case "windows":
    inFile = "C:\\Users\\WangFeng\\Desktop\\AndroidManifest.xml"
    outFile = "C:\\Users\\WangFeng\\Desktop\\AndroidManifest.json"
  case "linux":
    inFile = "/home/wangfeng/workspace/wechat/raw/AndroidManifest.xml"
    outFile = "/home/wangfeng/workspace/wechat/raw/AndroidManifest.json"
  default:
    panic(errors.New("os not supported"))
  }
  xml := ParseXml(inFile)
  if xml != nil {
    xml2 := NewXml2(xml)
    data, e := json.Marshal(xml2)
    if e != nil {
      panic(e)
    }
    e = ioutil.WriteFile(outFile, data, os.ModePerm)
    if e != nil {
      panic(e)
    }
  }
}
