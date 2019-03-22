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

  *Header
  StrPool    *StrPool
  ResId      *ResId
  Namespaces []*Namespace
  Tags       []*Tag
}

func (xml *Xml) writeTo(w *bytesWriter) {
  for i, o := range xml.o {
    switch v := o.(type) {
    case *Header:
      v.writeTo(w)
    case *StrPool:
      v.writeTo(w)
    case *ResId:
      v.writeTo(w)
    case *Namespace:
      v.writeTo(w)
    case *Tag:
      v.writeTo(w)
    default:
      fmt.Printf("Xml.writeTo(): unsupported type(%T)\n", v)
    }
    if i%100 == 0 {
      w.Flush()
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

func (xml *Xml) AddStr(str string) uint32 {
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

// value只能是string/int/bool
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
  case string:
    dataType = 3
    data = xml.AddStr(v)
    rawValue = data
    decodedValue = v
  case int:
    dataType = 16
    data = uint32(v)
    decodedValue = strconv.Itoa(v)
  case bool:
    dataType = 18
    decodedValue = "true"
    if !v {
      data = 0
      decodedValue = "false"
    }
  default:
    return errors.New(fmt.Sprintf("Xml.AddAttr(): unsupported type(%T)", v))
  }

  var decodedNamespacePrefix, decodedName string
  arr := strings.Split(key, ":")
  if len(arr) != 2 || arr[0] == "" {
    decodedName = key
  } else {
    decodedNamespacePrefix = arr[0]
    decodedName = arr[1]
  }

  for _, attr := range tag.Attrs {
    if attr.DecodedNamespacePrefix == decodedNamespacePrefix && attr.DecodedName == decodedName {
      if attr.DataType == 3 || attr.DataType == 16 || attr.DataType == 18 {
        attr.DataType = dataType
        attr.Data = data
        attr.RawValue = rawValue
        attr.DecodedValue = decodedValue
        if decodedNamespacePrefix == "" {
          attr.DecodedFull = attr.DecodedName + "=\"" + attr.DecodedValue + "\""
        } else {
          attr.DecodedFull = attr.DecodedNamespacePrefix + ":" + attr.DecodedName + "=\"" + attr.DecodedValue + "\""
        }
        return nil
      }
      return errors.New("attr already exists but its data type is not string/int/bool")
    }
  }

  namespaceUri := uint32(math.MaxUint32)
  name := xml.AddStr(decodedName)
  decodedFull := decodedName + "=\"" + decodedValue + "\""
  if decodedNamespacePrefix != "" {
    for k, p := range xml.p {
      if decodedNamespacePrefix == p {
        namespaceUri = k
        break
      }
    }
    if namespaceUri < math.MaxUint32 {
      decodedFull = decodedNamespacePrefix + ":" + decodedFull
    }
  }

  attr := &Attr{
    DecodedNamespacePrefix: decodedNamespacePrefix,
    DecodedName:            decodedName,
    DecodedValue:           decodedValue,
    DecodedFull:            decodedFull,
    NamespaceUri:           namespaceUri,
    Name:                   name,
    RawValue:               rawValue,
    ValueSize:              8,
    DataType:               dataType,
    Data:                   data,
  }
  tag.Attrs = append(tag.Attrs, attr)
  tag.AttrCount += 1
  tag.Size += 20
  xml.Size += 20
  return nil
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
  o := make([]interface{}, 0, 4096)

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
      fmt.Printf("DecodeXml(): unsupported type(%d)\n", v)
    }
  }

  ret := &Xml{
    o:          o,
    p:          make(map[uint32]string, 4),
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
  header := decodeHeader(r)
  ids := r.readUint32Array((header.Size - 8) / 4)
  return &ResId{
    Header: header,
    Ids:    ids,
  }
}

func decodeNamespace(r *bytesReader) *Namespace {
  return &Namespace{
    Header:     decodeHeader(r),
    LineNumber: r.readUint32(),
    Res0:       r.readUint32(),
    Prefix:     r.readUint32(),
    Uri:        r.readUint32(),
  }
}

func decodeTag(r *bytesReader) *Tag {
  header := decodeHeader(r)
  lineNumber := r.readUint32()
  res0 := r.readUint32()
  namespaceUri := r.readUint32()
  name := r.readUint32()

  if header.Type == 259 {
    return &Tag{
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
    NamespaceUri: xml.readUint32(),
    Name:         xml.readUint32(),
    RawValue:     xml.readUint32(),
    ValueSize:    xml.readUint16(),
    Res0:         xml.readUint8(),
    DataType:     xml.readUint8(),
    Data:         xml.readUint32(),
  }
}
