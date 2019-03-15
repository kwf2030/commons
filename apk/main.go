package main

import (
  "encoding/json"
  "io/ioutil"
  "os"
  "path"
)

func main() {
  // debugResTable()
  // debugManifest()
  setDebuggable(true)
}

func debugResTable() {
  rt := ParseResTable(path.Join("testdata", "resources.arsc"))
  if rt == nil {
    return
  }
  rt2 := NewResTable2(rt)
  data, e := json.Marshal(rt2)
  if e != nil {
    panic(e)
  }
  e = ioutil.WriteFile(path.Join("testdata", "resources.json"), data, os.ModePerm)
  if e != nil {
    panic(e)
  }
}

func debugManifest() {
  xml := ParseXml(path.Join("testdata", "AndroidManifest.xml"))
  if xml == nil {
    return
  }
  xml2 := NewXml2(xml)
  data, e := json.Marshal(xml2)
  if e != nil {
    panic(e)
  }
  e = ioutil.WriteFile(path.Join("testdata", "AndroidManifest.json"), data, os.ModePerm)
  if e != nil {
    panic(e)
  }
}

func setDebuggable(debuggable bool) {
  xml := ParseXml(path.Join("testdata", "AndroidManifest.xml"))
  if xml == nil {
    return
  }
  xml2 := NewXml2(xml)
  data, e := json.Marshal(xml2)
  if e != nil {
    panic(e)
  }
  e = ioutil.WriteFile(path.Join("testdata", "AndroidManifest.json"), data, os.ModePerm)
  if e != nil {
    panic(e)
  }

  /*str := "debuggable"
  index := uint32(math.MaxUint32)
  for i, s := range xml2.Ori.StrPool.Strs {
    if s == str {
      index = uint32(i)
      break
    }
  }*/
}
