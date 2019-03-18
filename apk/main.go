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
  xml := ParseXml(path.Join("testdata", "AndroidManifest2.xml"))
  if xml == nil {
    return
  }
  xml2 := NewXml2(xml)
  data, e := json.Marshal(xml2)
  if e != nil {
    panic(e)
  }
  e = ioutil.WriteFile(path.Join("testdata", "AndroidManifest2.json"), data, os.ModePerm)
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

  value := conv.Uint32ToBytesL(uint32(math.MaxUint32))
  if !debuggable {
    value = conv.Uint32ToBytesL(uint32(0))
  }

  // 定位到application节点
  var appTag *XmlTag2
  for _, tag2 := range xml2.Tags2 {
    if tag2.Name == "application" {
      appTag = tag2
      break
    }
  }
  // 定位到application节点的debuggable属性
  for i, attr := range appTag.Attrs {
    if strings.Contains(attr, "android:debuggable") {
      raw := xml2.Ori.bytesReader.data
      pos := appTag.Ori.Attrs[i].ChunkStart + 16
      buf := bytes.Buffer{}
      buf.Write(raw[:pos])
      buf.Write(value)
      buf.Write(raw[pos+4:])
      f, _ := os.OpenFile(path.Join("testdata", "AndroidManifest2.xml"), os.O_CREATE|os.O_TRUNC, os.ModePerm)
      buf.WriteTo(f)
      f.Close()
      return
    }
  }
  poolBytes := xml.AddStrToPool("debuggable")
  buf := bytes.Buffer{}
  buf.Write(xml.data[:xml.StrPool.ChunkStart])
  buf.Write(poolBytes)
  buf.Write(xml.data[xml.StrPool.ChunkEnd:])
  f, _ := os.OpenFile(path.Join("testdata", "AndroidManifest2.xml"), os.O_CREATE|os.O_TRUNC, os.ModePerm)
  buf.WriteTo(f)
  f.Close()
}
