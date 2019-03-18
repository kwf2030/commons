package main

import (
  "bytes"
  "io/ioutil"
  "math"
)

type XmlHeader struct {
  Type       uint16
  HeaderSize uint16
  Size       uint32
}

type XmlStrPool struct {
  ChunkStart, ChunkEnd uint32

  *XmlHeader
  StrCount     uint32
  StyleCount   uint32
  Flags        uint32
  StrStart     uint32
  StyleStart   uint32
  StrOffsets   []uint32
  StyleOffsets []uint32
  Strs         []string
  Styles       []byte
}

type XmlResId struct {
  ChunkStart, ChunkEnd uint32

  *XmlHeader
  Ids []uint32
}

type XmlNamespace struct {
  ChunkStart, ChunkEnd uint32

  *XmlHeader
  LineNumber uint32
  Res0       uint32
  Prefix     uint32
  Uri        uint32
}

type XmlTag struct {
  ChunkStart, ChunkEnd uint32

  *XmlHeader
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

type Xml struct {
  *bytesReader `json:"-"`

  ChunkStart, ChunkEnd uint32

  *XmlHeader
  // UTF-16
  StrPool    *XmlStrPool
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
  xml := &Xml{bytesReader: &bytesReader{Reader: bytes.NewReader(data), data: data}, ChunkStart: 0}
  xml.XmlHeader = xml.parseHeader()
  xml.StrPool = xml.parseStrPool()
  xml.ResId = xml.parseResId()
  xml.Namespaces = make([]*XmlNamespace, 0, 4)
  xml.Tags = make([]*XmlTag, 0, 4096)
  for xml.pos() < xml.Size {
    switch xml.readUint16() {
    case 258:
      xml.unreadN(2)
      xml.Tags = append(xml.Tags, xml.parseStartTag())
    case 259:
      xml.unreadN(2)
      xml.Tags = append(xml.Tags, xml.parseEndTag())
    case 256, 257:
      xml.unreadN(2)
      xml.Namespaces = append(xml.Namespaces, xml.parseNamespace())
    }
  }
  xml.ChunkEnd = xml.Size
  return xml
}

func (xml *Xml) parseHeader() *XmlHeader {
  return &XmlHeader{
    Type:       xml.readUint16(),
    HeaderSize: xml.readUint16(),
    Size:       xml.readUint32(),
  }
}

func (xml *Xml) parseStrPool() *XmlStrPool {
  chunkStart := xml.pos()
  header := xml.parseHeader()
  strCount := xml.readUint32()
  styleCount := xml.readUint32()
  flags := xml.readUint32()
  strStart := xml.readUint32()
  styleStart := xml.readUint32()
  strOffsets := xml.readUint32Array(strCount)
  styleOffsets := xml.readUint32Array(styleCount)

  var strs []string
  if strCount > 0 && styleCount < math.MaxUint32 {
    end := chunkStart + header.Size
    if styleCount > 0 && styleCount < math.MaxUint32 {
      end = chunkStart + styleStart
    }
    pool := xml.slice(xml.pos(), end)
    strs = make([]string, strCount)
    if flags&0x0100 != 0 {
      for i := uint32(0); i < strCount; i++ {
        strs[i] = string(str8(pool, strOffsets[i]))
      }
    } else {
      for i := uint32(0); i < strCount; i++ {
        b := str16(pool, strOffsets[i])
        arr := make([]byte, 0, len(b))
        for _, v := range b {
          if v != 0 {
            arr = append(arr, v)
          }
        }
        strs[i] = string(arr)
      }
    }
  }

  // todo 样式解析
  var styles []byte
  if styleCount > 0 && styleCount < math.MaxUint32 {
    styles = xml.slice(chunkStart+styleStart, chunkStart+header.Size)
  }

  return &XmlStrPool{
    ChunkStart:   chunkStart,
    ChunkEnd:     chunkStart + header.Size,
    XmlHeader:    header,
    StrCount:     strCount,
    StyleCount:   styleCount,
    Flags:        flags,
    StrStart:     strStart,
    StyleStart:   styleStart,
    StrOffsets:   strOffsets,
    StyleOffsets: styleOffsets,
    Strs:         strs,
    Styles:       styles,
  }
}

func (xml *Xml) parseResId() *XmlResId {
  chunkStart := xml.pos()
  header := xml.parseHeader()
  ids := xml.readUint32Array((header.Size - 8) / 4)
  return &XmlResId{
    ChunkStart: chunkStart,
    ChunkEnd:   chunkStart + header.Size,
    XmlHeader:  header,
    Ids:        ids,
  }
}

func (xml *Xml) parseNamespace() *XmlNamespace {
  chunkStart := xml.pos()
  header := xml.parseHeader()
  lineNumber := xml.readUint32()
  res0 := xml.readUint32()
  prefix := xml.readUint32()
  uri := xml.readUint32()
  return &XmlNamespace{
    ChunkStart: chunkStart,
    ChunkEnd:   chunkStart + header.Size,
    XmlHeader:  header,
    LineNumber: lineNumber,
    Res0:       res0,
    Prefix:     prefix,
    Uri:        uri,
  }
}

func (xml *Xml) parseStartTag() *XmlTag {
  chunkStart := xml.pos()
  header := xml.parseHeader()
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
    XmlHeader:    header,
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
  header := xml.parseHeader()
  return &XmlTag{
    ChunkStart:   chunkStart,
    ChunkEnd:     chunkStart + header.Size,
    XmlHeader:    header,
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
