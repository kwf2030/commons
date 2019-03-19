package apk

import (
  "encoding/json"
  "errors"
  "io/ioutil"
  "os"
  "path"
  "testing"
)

func TestManifest(t *testing.T) {
  name := path.Join("testdata", "AndroidManifest")

  m1, e := decodeManifestFromBinary(name)
  if e != nil {
    t.Fatal(e)
    return
  }

  e = encodeManifestToBinary(name+"_encode", m1)
  if e != nil {
    t.Fatal(e)
    return
  }
  m2, e := decodeManifestFromBinary(name + "_encode")
  if e != nil {
    t.Fatal(e)
    return
  }

  assertUint32Equals(t, m1.ChunkStart, m2.ChunkStart)
  assertUint32Equals(t, m1.ChunkEnd, m2.ChunkEnd)
  assertHeaderEquals(t, m1.XmlHeader, m2.XmlHeader)
  assertStrPoolEquals(t, m1.StrPool, m2.StrPool)
  assertResIdEquals(t, m1.ResId, m2.ResId)
  for i := 0; i < len(m1.Namespaces); i++ {
    // 最后一个Namespace错误，排查一下解析方法
    assertNamespaceEquals(t, m1.Namespaces[i], m2.Namespaces[i])
  }
  //for i := 0; i < len(m1.Tags); i++ {
  //  assertTagEquals(t, m1.Tags[i], m2.Tags[i])
  //}
}

func decodeManifestFromBinary(name string) (*Xml, error) {
  xml := ParseXml(name + ".xml")
  if xml == nil {
    return nil, errors.New("parse failed")
  }
  return xml, nil
}

func decodeManifestFromJson(name string) (*Xml, error) {
  data, e := ioutil.ReadFile(name + ".json")
  if e != nil {
    return nil, e
  }
  ret := &Xml{}
  e = json.Unmarshal(data, ret)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

func encodeManifestToBinary(name string, xml *Xml) error {
  f, e := os.OpenFile(name+".xml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
  if e != nil {
    return e
  }
  xml.writeTo(newBytesWriter(f))
  f.Close()
  return nil
}

func encodeManifestToJson(name string, xml *Xml) error {
  data, e := json.Marshal(xml)
  if e != nil {
    return e
  }
  e = ioutil.WriteFile(name+".json", data, os.ModePerm)
  if e != nil {
    return e
  }
  return nil
}

func assertHeaderEquals(t *testing.T, h1, h2 *XmlHeader) {
  assertUint16Equals(t, h1.Type, h2.Type)
  assertUint16Equals(t, h1.HeaderSize, h2.HeaderSize)
  assertUint32Equals(t, h1.Size, h2.Size)
}

func assertStrPoolEquals(t *testing.T, p1, p2 *XmlStrPool) {
  assertUint32Equals(t, p1.ChunkStart, p2.ChunkStart)
  assertUint32Equals(t, p1.ChunkEnd, p2.ChunkEnd)
  assertHeaderEquals(t, p1.XmlHeader, p2.XmlHeader)
  assertUint32Equals(t, p1.StrCount, p2.StrCount)
  assertUint32Equals(t, p1.StyleCount, p2.StyleCount)
  assertUint32Equals(t, p1.Flags, p2.Flags)
  assertUint32Equals(t, p1.StrStart, p2.StrStart)
  assertUint32Equals(t, p1.StyleStart, p2.StyleStart)
  assertUint32ArrayEquals(t, p1.StrOffsets, p2.StrOffsets)
  assertUint32ArrayEquals(t, p1.StyleOffsets, p2.StyleOffsets)
  assertStrArrayEquals(t, p1.Strs, p2.Strs)
  assertByteArrayEquals(t, p1.Styles, p2.Styles)
}

func assertResIdEquals(t *testing.T, r1, r2 *XmlResId) {
  assertUint32Equals(t, r1.ChunkStart, r2.ChunkStart)
  assertUint32Equals(t, r1.ChunkEnd, r2.ChunkEnd)
  assertHeaderEquals(t, r1.XmlHeader, r2.XmlHeader)
  assertUint32ArrayEquals(t, r1.Ids, r2.Ids)
}

func assertNamespaceEquals(t *testing.T, ns1, ns2 *XmlNamespace) {
  assertUint32Equals(t, ns1.ChunkStart, ns2.ChunkStart)
  assertUint32Equals(t, ns1.ChunkEnd, ns2.ChunkEnd)
  assertHeaderEquals(t, ns1.XmlHeader, ns2.XmlHeader)
  assertUint32Equals(t, ns1.LineNumber, ns2.LineNumber)
  assertUint32Equals(t, ns1.Res0, ns2.Res0)
  assertUint32Equals(t, ns1.Prefix, ns2.Prefix)
  assertUint32Equals(t, ns1.Uri, ns2.Uri)
}

func assertTagEquals(t *testing.T, t1, t2 *XmlTag) {
  assertUint32Equals(t, t1.ChunkStart, t2.ChunkStart)
  assertUint32Equals(t, t1.ChunkEnd, t2.ChunkEnd)
  assertHeaderEquals(t, t1.XmlHeader, t2.XmlHeader)
  assertUint32Equals(t, t1.LineNumber, t2.LineNumber)
  assertUint32Equals(t, t1.Res0, t2.Res0)
  assertUint32Equals(t, t1.NamespaceUri, t2.NamespaceUri)
  assertUint32Equals(t, t1.Name, t2.Name)
  assertUint16Equals(t, t1.AttrStart, t2.AttrStart)
  assertUint16Equals(t, t1.AttrSize, t2.AttrSize)
  assertUint16Equals(t, t1.AttrCount, t2.AttrCount)
  assertUint16Equals(t, t1.IdIndex, t2.IdIndex)
  assertUint16Equals(t, t1.ClassIndex, t2.ClassIndex)
  assertUint16Equals(t, t1.StyleIndex, t2.StyleIndex)
  for i := uint16(0); i < t1.AttrCount; i++ {
    assertAttrEquals(t, t1.Attrs[i], t2.Attrs[i])
  }
}

func assertAttrEquals(t *testing.T, a1, a2 *XmlAttr) {
  assertUint32Equals(t, a1.ChunkStart, a2.ChunkStart)
  assertUint32Equals(t, a1.ChunkEnd, a2.ChunkEnd)
  assertUint32Equals(t, a1.NamespaceUri, a2.NamespaceUri)
  assertUint32Equals(t, a1.Name, a2.Name)
  assertUint32Equals(t, a1.RawValue, a2.RawValue)
  assertUint16Equals(t, a1.ValueSize, a2.ValueSize)
  assertUint8Equals(t, a1.Res0, a2.Res0)
  assertUint8Equals(t, a1.DataType, a2.DataType)
  assertUint32Equals(t, a1.Data, a2.Data)
}

func assertUint8Equals(t *testing.T, n1, n2 uint8) {
  if n1 != n2 {
    t.Fatalf("assertUint8Equals failed, n1=%d, n2=%d", n1, n2)
  }
}

func assertUint16Equals(t *testing.T, n1, n2 uint16) {
  if n1 != n2 {
    t.Fatalf("assertUint16Equals failed, n1=%d, n2=%d", n1, n2)
  }
}

func assertUint32Equals(t *testing.T, n1, n2 uint32) {
  if n1 != n2 {
    t.Fatalf("assertUint32Equals failed, n1=%d, n2=%d", n1, n2)
  }
}

func assertByteArrayEquals(t *testing.T, arr1, arr2 []byte) {
  l1, l2 := len(arr1), len(arr2)
  if l1 != l2 {
    t.Fatalf("assertByteArrayEquals failed, len(arr1)=%d, len(arr2)=%d", l1, l2)
    return
  }
  for i := 0; i < l1; i++ {
    if arr1[i] != arr2[i] {
      t.Fatalf("assertByteArrayEquals failed, arr1[%d]=%d, arr2[%d]=%d", i, arr1[i], i, arr2[i])
    }
  }
}

func assertStrArrayEquals(t *testing.T, arr1, arr2 []string) {
  l1, l2 := len(arr1), len(arr2)
  if l1 != l2 {
    t.Fatalf("assertStrArrayEquals failed, len(arr1)=%d, len(arr2)=%d", l1, l2)
    return
  }
  for i := 0; i < l1; i++ {
    if arr1[i] != arr2[i] {
      t.Fatalf("assertStrArrayEquals failed, arr1[%d]=%s, arr2[%d]=%s", i, arr1[i], i, arr2[i])
    }
  }
}

func assertUint32ArrayEquals(t *testing.T, arr1, arr2 []uint32) {
  l1, l2 := len(arr1), len(arr2)
  if l1 != l2 {
    t.Fatalf("assertUint32ArrayEquals failed, len(arr1)=%d, len(arr2)=%d", l1, l2)
    return
  }
  for i := 0; i < l1; i++ {
    if arr1[i] != arr2[i] {
      t.Fatalf("assertUint32ArrayEquals failed, arr1[%d]=%d, arr2[%d]=%d", i, arr1[i], i, arr2[i])
    }
  }
}

/*func setDebuggable(debuggable bool) {
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
  // f, _ := os.OpenFile(path.Join("testdata", "AndroidManifest2.xml"), os.O_CREATE|os.O_TRUNC, os.ModePerm)
  // f.Write(xml.Encode())
  // f.Close()
}*/
