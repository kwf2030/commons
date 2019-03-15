package main

import (
  "encoding/json"
  "io/ioutil"
  "os"
  "path"
)

var dir = path.Join(os.Getenv("GOPATH"), "src", "github.com", "kwf2030", "commons", "apk", "testdata")

func main() {
  // debugResTable()
  // debugManifest()
  setDebuggable(true)
}

func debugResTable() {
  rt := ParseResTable(path.Join(dir, "resources.arsc"))
  if rt == nil {
    return
  }
  rt2 := NewResTable2(rt)
  data, e := json.Marshal(rt2)
  if e != nil {
    panic(e)
  }
  e = ioutil.WriteFile(path.Join(dir, "resources.json"), data, os.ModePerm)
  if e != nil {
    panic(e)
  }
}

func debugManifest() {
  xml := ParseXml(path.Join(dir, "AndroidManifest.xml"))
  if xml == nil {
    return
  }
  xml2 := NewXml2(xml)
  data, e := json.Marshal(xml2)
  if e != nil {
    panic(e)
  }
  e = ioutil.WriteFile(path.Join(dir, "AndroidManifest.json"), data, os.ModePerm)
  if e != nil {
    panic(e)
  }
}

func setDebuggable(debuggable bool) {
  xml := ParseXml(path.Join(dir, "AndroidManifest.xml"))
  if xml == nil {
    return
  }
  xml2 := NewXml2(xml)
  data, e := json.Marshal(xml2)
  if e != nil {
    panic(e)
  }
  e = ioutil.WriteFile(path.Join(dir, "AndroidManifest.json"), data, os.ModePerm)
  if e != nil {
    panic(e)
  }

  /*for _, tag2 := range xml2.Tags2 {
  }*/
}
