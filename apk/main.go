package main

import (
  "bytes"
  "encoding/json"
  "io/ioutil"
  "math"
  "os"
  "path"
  "strings"

  "github.com/kwf2030/commons/conv"
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

  // 定位到application节点
  var appTag *XmlTag2
  for _, tag2 := range xml2.Tags2 {
    if tag2.Name == "application" {
      appTag = tag2
      break
    }
  }
  // 定位到application节点的debuggable属性，修改值为0xFFFFFFFF（true）
  for i, attr := range appTag.Attrs {
    if strings.Contains(attr, "android:debuggable") {
      raw := xml2.Ori.bytesReader.data
      pos := appTag.Ori.Attrs[i].ChunkStart + 16
      buf := bytes.Buffer{}
      buf.Write(raw[:pos])
      buf.Write(conv.Uint32ToBytesL(uint32(math.MaxUint32)))
      buf.Write(raw[pos+4:])
      f, _ := os.OpenFile(path.Join("testdata", "AndroidManifest2.xml"), os.O_CREATE|os.O_TRUNC, os.ModePerm)
      buf.WriteTo(f)
      return
    }
  }
  // 查找字符串池有没有debuggable
  /*index := uint32(math.MaxUint32)
  for i, s := range xml2.Ori.StrPool.Strs {
    if s == "debuggable" {
      index = uint32(i)
      break
    }
  }*/
}
