package apk

import (
  "encoding/json"
  "errors"
  "fmt"
  "io/ioutil"
  "math"
  "os"
  "strconv"
  "strings"
)

type Xml struct {
  o []interface{}

  p map[uint32]string

  ChunkStart, ChunkEnd uint32

  *Header
  StrPool    *StrPool
  ResId      *ResId
  Namespaces []*Namespace
  Tags       []*Tag
}

func (xml *Xml) writeTo(w *bytesWriter) {
  xml.Header.writeTo(w)
  w.Flush()
  xml.StrPool.writeTo(w)
  w.Flush()
  xml.ResId.writeTo(w)
  w.Flush()
  // 同一个struct数组是有序的（按解析顺序），但不同的struct数组没有记录顺序，
  // 如Xml.Namespaces和Xml.Tags两个数组，各自本身是按解析顺序存储的，
  // 但两个数组在解析时是交叉的（Namespace->Tag->Tag->Namespace），
  // 实际上是先一个Namespace，然后全部的Tag，最后再一个Namespace（通常就是文件的结束），
  // 注意，这种情况只有不同的struct数组交叉解析才会出现，
  // 且只影响写入顺序，不影响解析结果（解析出来的ChunkStart/ChunkEnd字段和原来不同），
  // 因为struct的ChunkStart/ChunkEnd字段可以表示其读取顺序，
  // 所以这里用其来保证写入的顺序和读取的顺序一致
  last := xml.ResId.ChunkEnd
  for _, ns := range xml.Namespaces {
    if ns.ChunkStart == last {
      last = ns.ChunkEnd
      ns.writeTo(w)
    }
  }
  w.Flush()
  for i, t := range xml.Tags {
    if t.ChunkStart == last {
      last = t.ChunkEnd
      t.writeTo(w)
    }
    if i%100 == 0 {
      w.Flush()
    }
  }
  w.Flush()
  for _, ns := range xml.Namespaces {
    if ns.ChunkStart == last {
      last = ns.ChunkEnd
      ns.writeTo(w)
    }
  }
  w.Flush()
}

func (xml *Xml) parseNamespaces() {
  for _, ns := range xml.Namespaces {
    if ns.Prefix < math.MaxUint32 {
      ns.DecodedPrefix = xml.StrPool.Strs[ns.Prefix]
      xml.p[ns.Uri] = ns.DecodedPrefix
    }
    if ns.Uri < math.MaxUint32 {
      ns.DecodedUri = xml.StrPool.Strs[ns.Uri]
    }
  }
}

func (xml *Xml) parseTags() {
  for _, tag := range xml.Tags {
    tag.DecodedName = xml.StrPool.Strs[tag.Name]
    tag.DecodedFull = tag.DecodedName
    if tag.NamespaceUri < math.MaxUint32 {
      tag.DecodedNamespacePrefix = xml.p[tag.NamespaceUri]
      tag.DecodedFull = tag.DecodedNamespacePrefix + ":" + tag.DecodedName
    }
    for _, attr := range tag.Attrs {
      attr.DecodedName = xml.StrPool.Strs[attr.Name]
      attr.DecodedFull = attr.DecodedName
      if attr.NamespaceUri < math.MaxUint32 {
        attr.DecodedNamespacePrefix = xml.p[attr.NamespaceUri]
        attr.DecodedFull = attr.DecodedNamespacePrefix + ":" + attr.DecodedName
      }
      attr.DecodedValue = xml.parseData(attr.DataType, attr.Data)
      attr.DecodedFull += "=\"" + attr.DecodedValue + "\""
    }
  }
}

func (xml *Xml) parseData(dataType uint8, data uint32) string {
  switch dataType {
  case 3:
    return xml.StrPool.Strs[data]
  case 16:
    return strconv.FormatUint(uint64(data), 10)
  case 18:
    if data == 0 {
      return "false"
    }
    return "true"
  }
  return ""
}

func (xml *Xml) addStr(str string) uint32 {
  if str == "" {
    return math.MaxUint32
  }
  pool := xml.StrPool
  for i, v := range pool.Strs {
    if v == str {
      return uint32(i)
    }
  }

  strLen := uint32(2 + 2*len(str) + 2)
  size := pool.Size + 4 + strLen
  strCount := pool.StrCount + 1
  strStart := pool.StrStart + 4
  styleStart := pool.StyleStart + 4 + strLen
  lastStrLen := uint32(2 + 2*len(pool.Strs[pool.StrCount-1]) + 2)
  strOffset := pool.StrOffsets[pool.StrCount-1] + lastStrLen

  pool.Size = size
  pool.StrCount = strCount
  pool.StrStart = strStart
  if pool.StyleCount > 0 {
    pool.StyleStart = styleStart
  }
  pool.StrOffsets = append(pool.StrOffsets, strOffset)
  pool.Strs = append(pool.Strs, str)
  xml.Size += 4 + strLen

  return strCount - 1
}

// value只能是数字/字符串/布尔值
func (xml *Xml) AddAttr(key string, value interface{}, f func(*Tag) bool) error {
  var tag *Tag
  for _, t := range xml.Tags {
    if f(t) {
      tag = t
      break
    }
  }
  if tag == nil {
    return errors.New("tag not found")
  }

  dataType, data, rawValue := uint8(math.MaxUint8), uint32(math.MaxUint32), uint32(math.MaxUint32)
  var decodedValue string
  switch v := value.(type) {
  case bool:
    dataType = 18
    decodedValue = "true"
    if !v {
      data = 0
      decodedValue = "false"
    }
  case int:
    dataType = 16
    data = uint32(v)
    decodedValue = strconv.Itoa(v)
  case string:
    dataType = 3
    data = xml.addStr(v)
    rawValue = data
    decodedValue = v
  }

  arr := strings.Split(key, ":")
  var prefix, name string
  if len(arr) != 2 || arr[0] == "" {
    name = key
  } else {
    prefix = arr[0]
    name = arr[1]
  }

  for _, attr := range tag.Attrs {
    if attr.DecodedNamespacePrefix == prefix && attr.DecodedName == name {
      // todo 修改attr的Data
      return nil
    }
  }

  return nil

  /*


  for _, attr := range tag.Attrs {

  }

  attr := &Attr{
    NamespaceUri: math.MaxUint32,
    Name:         math.MaxUint32,
    RawValue:     math.MaxUint32,
    ValueSize:    8,
    DataType:     math.MaxUint8,
    Data:         math.MaxUint32,
  }



  if arr := strings.Split(key, ":"); len(arr) != 2 || arr[0] == "" {
    attr.Name = xml.addStr(key)
    attr.DecodedName = xml.StrPool.Strs[attr.Name]
    attr.DecodedFull = attr.DecodedName
  } else {
    for k, prefix := range xml.p {
      if arr[0] == prefix {
        attr.NamespaceUri = k
        break
      }
    }
    if attr.NamespaceUri < math.MaxUint32 {
      attr.Name = xml.addStr(arr[1])
      attr.DecodedName = xml.StrPool.Strs[attr.Name]
      attr.DecodedNamespacePrefix = xml.p[attr.NamespaceUri]
      attr.DecodedFull = attr.DecodedNamespacePrefix + ":" + attr.DecodedName
    } else {
      attr.Name = xml.addStr(key)
      attr.DecodedName = xml.StrPool.Strs[attr.Name]
      attr.DecodedFull = attr.DecodedName
    }
  }

  attr.DecodedFull += "=\"" + attr.DecodedValue + "\""
  attr.ChunkStart = tag.Attrs[tag.AttrCount-1].ChunkEnd
  attr.ChunkEnd = attr.ChunkStart + 20
  tag.Attrs = append(tag.Attrs, attr)
  tag.AttrCount += 1
  tag.Size += 20
  xml.Size += 20
  return nil*/
}

func (xml *Xml) Marshal(name string) error {
  f, e := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
  if e != nil {
    return e
  }
  xml.writeTo(newBytesWriter(f))
  f.Close()
  return nil
}

func (xml *Xml) MarshalJSON(name string) error {
  data, e := json.Marshal(xml)
  if e != nil {
    return e
  }
  return ioutil.WriteFile(name, data, os.ModePerm)
}

type ResId struct {
  ChunkStart, ChunkEnd uint32

  *Header
  Ids []uint32
}

func (r *ResId) writeTo(w *bytesWriter) {
  r.Header.writeTo(w)
  w.writeUint32Array(r.Ids)
}

type Namespace struct {
  DecodedPrefix string
  DecodedUri    string

  ChunkStart, ChunkEnd uint32

  *Header
  LineNumber uint32
  Res0       uint32
  Prefix     uint32
  Uri        uint32
}

func (ns *Namespace) writeTo(w *bytesWriter) {
  ns.Header.writeTo(w)
  w.writeUint32(ns.LineNumber)
  w.writeUint32(ns.Res0)
  w.writeUint32(ns.Prefix)
  w.writeUint32(ns.Uri)
}

type Tag struct {
  DecodedNamespacePrefix string
  DecodedName            string
  DecodedFull            string

  ChunkStart, ChunkEnd uint32

  *Header
  LineNumber   uint32
  Res0         uint32
  NamespaceUri uint32
  Name         uint32
  AttrStart    uint16
  AttrSize     uint16
  AttrCount    uint16
  IdIndex      uint16
  ClassIndex   uint16
  StyleIndex   uint16
  Attrs        []*Attr
}

func (t *Tag) writeTo(w *bytesWriter) {
  t.Header.writeTo(w)
  w.writeUint32(t.LineNumber)
  w.writeUint32(t.Res0)
  w.writeUint32(t.NamespaceUri)
  w.writeUint32(t.Name)
  if t.Type == 258 {
    w.writeUint16(t.AttrStart)
    w.writeUint16(t.AttrSize)
    w.writeUint16(t.AttrCount)
    w.writeUint16(t.IdIndex)
    w.writeUint16(t.ClassIndex)
    w.writeUint16(t.StyleIndex)
    for _, attr := range t.Attrs {
      attr.writeTo(w)
    }
  }
}

type Attr struct {
  DecodedNamespacePrefix string
  DecodedName            string
  DecodedValue           string
  DecodedFull            string

  ChunkStart, ChunkEnd uint32

  NamespaceUri uint32
  Name         uint32
  RawValue     uint32
  ValueSize    uint16
  Res0         uint8
  DataType     uint8
  Data         uint32
}

func (a *Attr) writeTo(w *bytesWriter) {
  w.writeUint32(a.NamespaceUri)
  w.writeUint32(a.Name)
  w.writeUint32(a.RawValue)
  w.writeUint16(a.ValueSize)
  w.writeUint8(a.Res0)
  w.writeUint8(a.DataType)
  w.writeUint32(a.Data)
}

func DecodeXml(file string) (*Xml, error) {
  if file == "" {
    return nil, errors.New("missing args")
  }
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return nil, e
  }
  r := newBytesReader(data)
  o := make([]interface{}, 4096)

  header := decodeHeader(r)
  o = append(o, header)

  strPool := decodeStrPool(r)
  o = append(o, strPool)

  resId := decodeResId(r)
  o = append(o, resId)

  nss := make([]*Namespace, 0, 4)
  tags := make([]*Tag, 0, 4096)
  for r.pos() < header.Size {
    switch v := r.readUint16(); v {
    case 258, 259:
      r.unreadN(2)
      t := decodeTag(r)
      tags = append(tags, t)
      o = append(o, t)
    case 256, 257:
      r.unreadN(2)
      ns := decodeNamespace(r)
      nss = append(nss, ns)
      o = append(o, ns)
    default:
      fmt.Println("unsupported tag type:", v)
    }
  }

  ret := &Xml{
    o:          o,
    p:          make(map[uint32]string, 4),
    ChunkStart: 0,
    ChunkEnd:   header.Size,
    Header:     header,
    StrPool:    strPool,
    ResId:      resId,
    Namespaces: nss,
    Tags:       tags,
  }
  ret.parseNamespaces()
  ret.parseTags()
  return ret, nil
}

func decodeResId(r *bytesReader) *ResId {
  chunkStart := r.pos()
  header := decodeHeader(r)
  ids := r.readUint32Array((header.Size - 8) / 4)
  return &ResId{
    ChunkStart: chunkStart,
    ChunkEnd:   chunkStart + header.Size,
    Header:     header,
    Ids:        ids,
  }
}

func decodeNamespace(r *bytesReader) *Namespace {
  chunkStart := r.pos()
  header := decodeHeader(r)
  return &Namespace{
    ChunkStart: chunkStart,
    ChunkEnd:   chunkStart + header.Size,
    Header:     header,
    LineNumber: r.readUint32(),
    Res0:       r.readUint32(),
    Prefix:     r.readUint32(),
    Uri:        r.readUint32(),
  }
}

func decodeTag(r *bytesReader) *Tag {
  chunkStart := r.pos()
  header := decodeHeader(r)
  lineNumber := r.readUint32()
  res0 := r.readUint32()
  namespaceUri := r.readUint32()
  name := r.readUint32()

  if header.Type == 259 {
    return &Tag{
      ChunkStart:   chunkStart,
      ChunkEnd:     chunkStart + header.Size,
      Header:       header,
      LineNumber:   lineNumber,
      Res0:         res0,
      NamespaceUri: namespaceUri,
      Name:         name,
    }
  }

  attrStart := r.readUint16()
  attrSize := r.readUint16()
  attrCount := r.readUint16()
  idIndex := r.readUint16()
  classIndex := r.readUint16()
  styleIndex := r.readUint16()

  var attrs []*Attr
  if uint16Valid(attrCount) {
    attrs = make([]*Attr, attrCount)
    for i := uint16(0); i < attrCount; i++ {
      attrs[i] = decodeAttr(r)
    }
  }

  return &Tag{
    ChunkStart:   chunkStart,
    ChunkEnd:     chunkStart + header.Size,
    Header:       header,
    LineNumber:   lineNumber,
    Res0:         res0,
    NamespaceUri: namespaceUri,
    Name:         name,
    AttrStart:    attrStart,
    AttrSize:     attrSize,
    AttrCount:    attrCount,
    IdIndex:      idIndex,
    ClassIndex:   classIndex,
    StyleIndex:   styleIndex,
    Attrs:        attrs,
  }
}

func decodeAttr(xml *bytesReader) *Attr {
  return &Attr{
    ChunkStart:   xml.pos(),
    NamespaceUri: xml.readUint32(),
    Name:         xml.readUint32(),
    RawValue:     xml.readUint32(),
    ValueSize:    xml.readUint16(),
    Res0:         xml.readUint8(),
    DataType:     xml.readUint8(),
    Data:         xml.readUint32(),
    ChunkEnd:     xml.pos(),
  }
}
