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
  Others     [][]byte
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
    case []byte:
      w.Write(v)
    default:
      panic(fmt.Sprintf("unsupported type: %T\n", v))
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

func (xml *Xml) AddStr(str string, i int) uint32 {
  if str == "" {
    return math.MaxUint32
  }

  pool := xml.StrPool
  for j, s := range pool.Strs {
    if s == str {
      return uint32(j)
    }
  }

  if i < -1 {
    i = 0
  } else if i >= len(xml.StrPool.Strs) {
    i = -1
  }
  u := uint32(i)

  strLen := uint32(2 + 2*len(str) + 2)
  size := pool.Size + 4 + strLen
  strCount := pool.StrCount + 1
  strStart := pool.StrStart + 4
  styleStart := pool.StyleStart + 4 + strLen
  strOffsets := make([]uint32, strCount)
  strs := make([]string, strCount)
  if i == -1 {
    lastStrLen := uint32(2 + 2*len(pool.Strs[pool.StrCount-1]) + 2)
    strOffset := pool.StrOffsets[pool.StrCount-1] + lastStrLen
    copy(strOffsets, pool.StrOffsets)
    strOffsets[strCount-1] = strOffset
    copy(strs, pool.Strs)
    strs[strCount-1] = str
  } else {
    copy(strOffsets, pool.StrOffsets[:u+1])
    copy(strs, pool.Strs[:u])
    strs[u] = str
    for j := u + 1; j < strCount; j++ {
      strOffsets[j] = pool.StrOffsets[j-1] + strLen
      strs[j] = pool.Strs[j-1]
    }
    for _, v := range xml.Namespaces {
      if v.Prefix >= u && v.Prefix < math.MaxUint32 {
        v.Prefix += 1
      }
      if v.Uri >= u && v.Uri < math.MaxUint32 {
        v.Uri += 1
      }
    }
    for _, t := range xml.Tags {
      if t.NamespaceUri >= u && t.NamespaceUri < math.MaxUint32 {
        t.NamespaceUri += 1
      }
      if t.Name >= u && t.Name < math.MaxUint32 {
        t.Name += 1
      }
      for _, a := range t.Attrs {
        if a.NamespaceUri >= u && a.NamespaceUri < math.MaxUint32 {
          a.NamespaceUri += 1
        }
        if a.Name >= u && a.Name < math.MaxUint32 {
          a.Name += 1
        }
        if a.RawValue >= u && a.RawValue < math.MaxUint32 {
          a.RawValue += 1
        }
        if a.DataType == 3 && a.Data >= u && a.Data < math.MaxUint32 {
          a.Data += 1
        }
      }
    }
    m := make(map[uint32]string, len(xml.p))
    for k, v := range xml.p {
      if k >= u && k < math.MaxUint32 {
        m[k+1] = v
      } else {
        m[k] = v
      }
    }
    xml.p = m
  }

  pool.Size = size
  pool.StrCount = strCount
  pool.StrStart = strStart
  if pool.StyleCount > 0 {
    pool.StyleStart = styleStart
  }
  pool.StrOffsets = strOffsets
  pool.Strs = strs
  xml.Size += 4 + strLen

  if i == -1 {
    return strCount - 1
  }
  return u
}

func (xml *Xml) AddResId(id uint32, i int) uint32 {
  for j, v := range xml.ResId.Ids {
    if v == id {
      return uint32(j)
    }
  }

  if i < -1 {
    i = 0
  } else if i >= len(xml.ResId.Ids) {
    i = -1
  }

  xml.ResId.Size += 4
  xml.Size += 4
  if i == -1 {
    xml.ResId.Ids = append(xml.ResId.Ids, id)
    return uint32(len(xml.ResId.Ids)) - 1
  } else {
    l := len(xml.ResId.Ids) + 1
    a := make([]uint32, l)
    copy(a, xml.ResId.Ids[:i])
    a[i] = id
    for j := i + 1; j < l; j++ {
      a[j] = xml.ResId.Ids[j-1]
    }
    xml.ResId.Ids = a
    return uint32(i)
  }
}

// value只能是string/int/bool
func (xml *Xml) AddAttr(key string, value interface{}, ai, ki, vi int, f func(*Tag) bool) error {
  if key == "" || value == nil || f == nil {
    return errors.New("Xml.AddAttr(): invalid args")
  }

  var tag *Tag
  for _, t := range xml.Tags {
    if f(t) {
      tag = t
      break
    }
  }
  if tag == nil {
    return errors.New("Xml.AddAttr(): tag not found")
  }

  if ai < -1 {
    ai = 0
  } else if ai >= len(tag.Attrs) {
    ai = -1
  }
  if ki < -1 {
    ki = 0
  } else if ki >= len(xml.StrPool.Strs) {
    ki = -1
  }
  if vi < -1 {
    vi = 0
  } else if vi >= len(xml.StrPool.Strs) {
    vi = -1
  }

  dataType, data, rawValue := uint8(math.MaxUint8), uint32(math.MaxUint32), uint32(math.MaxUint32)
  var decodedValue string
  switch v := value.(type) {
  case string:
    dataType = 3
    data = xml.AddStr(v, vi)
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
    return errors.New(fmt.Sprintf("Xml.AddAttr(): unsupported type: %T", v))
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
      return errors.New("Xml.AddAttr(): attr already exists but its data type is not string/int/bool")
    }
  }

  namespaceUri := uint32(math.MaxUint32)
  name := xml.AddStr(decodedName, ki)
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
  tag.Size += 20
  tag.AttrCount += 1
  if ai == -1 {
    tag.Attrs = append(tag.Attrs, attr)
  } else {
    l := len(tag.Attrs) + 1
    a := make([]*Attr, l)
    copy(a, tag.Attrs[:ai])
    a[ai] = attr
    for j := ai + 1; j < l; j++ {
      a[j] = tag.Attrs[j-1]
    }
    tag.Attrs = a
  }
  xml.Size += 20
  return nil
}

func (xml *Xml) Marshal(name string) error {
  if name == "" {
    return errors.New("Xml.Marshal(): invalid args")
  }
  f, e := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
  if e != nil {
    return e
  }
  xml.writeTo(newBytesWriter(f))
  f.Close()
  return nil
}

func (xml *Xml) MarshalJSON(name string) error {
  if name == "" {
    return errors.New("Xml.MarshalJSON(): invalid args")
  }
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

func DecodeXmlFile(file string) (*Xml, error) {
  if file == "" {
    return nil, errors.New("DecodeXmlFile(): invalid args")
  }
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return nil, e
  }
  return DecodeXml(data)
}

func DecodeXml(data []byte) (*Xml, error) {
  if len(data) == 0 {
    return nil, errors.New("DecodeXml(): invalid args")
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
  others := make([][]byte, 0, 64)
  for r.pos() < header.Size {
    switch v := r.readUint16(); v {
    case 256, 257:
      r.unreadN(2)
      ns := decodeNamespace(r)
      nss = append(nss, ns)
      o = append(o, ns)
    case 258, 259:
      r.unreadN(2)
      t := decodeTag(r)
      tags = append(tags, t)
      o = append(o, t)
    default:
      headerSize := uint32(r.readUint16())
      if headerSize < 4 || headerSize >= header.Size {
        panic(fmt.Sprintf("DecodeXml(): unsupported type: %d\n", v))
      }
      if headerSize == 4 {
        r.unreadN(4)
        b := r.readN(4)
        others = append(others, b)
        o = append(o, b)
      } else {
        s := r.readUint32()
        r.unreadN(8)
        b := r.readN(s)
        others = append(others, b)
        o = append(o, b)
      }
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
    Others:     others,
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
  if attrCount > 0 && attrCount < math.MaxUint16 {
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
