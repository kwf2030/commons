package apk

import (
  "fmt"
  "io/ioutil"
  "math"
)

type XmlResId struct {
  ChunkStart, ChunkEnd uint32

  *Header
  Ids []uint32
}

func (r *XmlResId) writeTo(w *bytesWriter) {
  r.Header.writeTo(w)
  w.writeUint32Array(r.Ids)
}

type XmlNamespace struct {
  ChunkStart, ChunkEnd uint32

  *Header
  LineNumber uint32
  Res0       uint32
  Prefix     uint32
  Uri        uint32
}

func (ns *XmlNamespace) writeTo(w *bytesWriter) {
  ns.Header.writeTo(w)
  w.writeUint32(ns.LineNumber)
  w.writeUint32(ns.Res0)
  w.writeUint32(ns.Prefix)
  w.writeUint32(ns.Uri)
}

type XmlTag struct {
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
  Attrs        []*XmlAttr
}

func (t *XmlTag) writeTo(w *bytesWriter) {
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

type XmlAttr struct {
  ChunkStart, ChunkEnd uint32

  NamespaceUri uint32
  Name         uint32
  RawValue     uint32
  ValueSize    uint16
  Res0         uint8
  DataType     uint8
  Data         uint32
}

func (a *XmlAttr) writeTo(w *bytesWriter) {
  w.writeUint32(a.NamespaceUri)
  w.writeUint32(a.Name)
  w.writeUint32(a.RawValue)
  w.writeUint16(a.ValueSize)
  w.writeUint8(a.Res0)
  w.writeUint8(a.DataType)
  w.writeUint32(a.Data)
}

type Xml struct {
  *bytesReader `json:"-"`

  ChunkStart, ChunkEnd uint32

  *Header
  // UTF-16
  StrPool    *StrPool
  ResId      *XmlResId
  Namespaces []*XmlNamespace
  Tags       []*XmlTag
}

func ParseXml(file string) *Xml {
  if file == "" {
    return nil
  }
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return nil
  }
  xml := &Xml{bytesReader: newBytesReader(data), ChunkStart: 0}
  xml.Header = parseHeader(xml.bytesReader)
  xml.StrPool = parseStrPool(xml.bytesReader)
  xml.ResId = xml.parseResId()
  xml.Namespaces = make([]*XmlNamespace, 0, 4)
  xml.Tags = make([]*XmlTag, 0, 4096)
  for xml.pos() < xml.Size {
    switch v := xml.readUint16(); v {
    case 258:
      xml.unreadN(2)
      xml.Tags = append(xml.Tags, xml.parseStartTag())
    case 259:
      xml.unreadN(2)
      xml.Tags = append(xml.Tags, xml.parseEndTag())
    case 256, 257:
      xml.unreadN(2)
      xml.Namespaces = append(xml.Namespaces, xml.parseNamespace())
    default:
      fmt.Println("unsupported tag:", v)
    }
  }
  xml.ChunkEnd = xml.Size
  return xml
}

func (xml *Xml) parseResId() *XmlResId {
  chunkStart := xml.pos()
  header := parseHeader(xml.bytesReader)
  ids := xml.readUint32Array((header.Size - 8) / 4)
  return &XmlResId{
    ChunkStart: chunkStart,
    ChunkEnd:   chunkStart + header.Size,
    Header:     header,
    Ids:        ids,
  }
}

func (xml *Xml) parseNamespace() *XmlNamespace {
  chunkStart := xml.pos()
  header := parseHeader(xml.bytesReader)
  lineNumber := xml.readUint32()
  res0 := xml.readUint32()
  prefix := xml.readUint32()
  uri := xml.readUint32()
  return &XmlNamespace{
    ChunkStart: chunkStart,
    ChunkEnd:   chunkStart + header.Size,
    Header:     header,
    LineNumber: lineNumber,
    Res0:       res0,
    Prefix:     prefix,
    Uri:        uri,
  }
}

func (xml *Xml) parseStartTag() *XmlTag {
  chunkStart := xml.pos()
  header := parseHeader(xml.bytesReader)
  lineNumber := xml.readUint32()
  res0 := xml.readUint32()
  namespaceUri := xml.readUint32()
  name := xml.readUint32()
  attrStart := xml.readUint16()
  attrSize := xml.readUint16()
  attrCount := xml.readUint16()
  idIndex := xml.readUint16()
  classIndex := xml.readUint16()
  styleIndex := xml.readUint16()

  var attrs []*XmlAttr
  if attrCount > 0 && attrCount < math.MaxUint16 {
    attrs = make([]*XmlAttr, attrCount)
    for i := uint16(0); i < attrCount; i++ {
      attrs[i] = xml.parseAttr()
    }
  }

  return &XmlTag{
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

func (xml *Xml) parseEndTag() *XmlTag {
  chunkStart := xml.pos()
  header := parseHeader(xml.bytesReader)
  return &XmlTag{
    ChunkStart:   chunkStart,
    ChunkEnd:     chunkStart + header.Size,
    Header:       header,
    LineNumber:   xml.readUint32(),
    Res0:         xml.readUint32(),
    NamespaceUri: xml.readUint32(),
    Name:         xml.readUint32(),
  }
}

func (xml *Xml) parseAttr() *XmlAttr {
  return &XmlAttr{
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
