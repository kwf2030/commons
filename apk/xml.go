package main

import (
  "bytes"
  "io"
  "io/ioutil"
  "math"
)

type XmlHeader struct {
  Type       uint16
  HeaderSize uint16
  Size       uint32
}

type XmlStrPool struct {
  *XmlHeader
  StrCount     uint32
  StyleCount   uint32
  Flags        uint32
  StrStart     uint32
  StyleStart   uint32
  StrOffsets   []uint32
  StyleOffsets []uint32
  Strs         []string
  Styles       []string
}

type XmlResId struct {
  *XmlHeader
  Ids []uint32
}

type XmlNamespace struct {
  *XmlHeader
  LineNumber uint32
  Res0       uint32
  Prefix     uint32
  Uri        uint32
}

type XmlTag struct {
  *XmlHeader
  LineNumber   uint32
  Res0         uint32
  NamespaceUri uint32
  Name         uint32
  Flags        uint32
  AttrCount    uint32
  ClassAttr    uint32
  Attrs        []*XmlAttr
}

type XmlAttr struct {
  Namespace uint32
  Uri       uint32
  Name      uint32
  Value     uint32
  Data      uint32
}

type Xml struct {
  *bytesReader
  *XmlHeader
  StrPool   *XmlStrPool
  ResId     *XmlResId
  Namespace []*XmlNamespace
  Tags      []*XmlTag
}

func ParseXml(file string) *Xml {
  if file == "" {
    return nil
  }
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return nil
  }
  xml := &Xml{bytesReader: &bytesReader{Reader: bytes.NewReader(data), data: data}}
  xml.XmlHeader = xml.parseXmlHeader()
  xml.StrPool = xml.parseXmlStrPool()
  xml.ResId = xml.parseXmlResId()
  xml.Namespace = make([]*XmlNamespace, 0, 4)
  xml.Tags = make([]*XmlTag, 0, 4096)
  for xml.pos() < xml.Size {
    switch xml.readUint16() {
    case 0x102:
      xml.unreadN(2)
      xml.Tags = append(xml.Tags, xml.parseXmlStartTag())
    case 0x103:
      xml.unreadN(2)
      xml.Tags = append(xml.Tags, xml.parseXmlEndTag())
    case 0x0100, 0x0101:
      xml.unreadN(2)
      xml.Namespace = append(xml.Namespace, xml.parseXmlNamespace())
    }
  }
  return xml
}

func (xml *Xml) parseXmlHeader() *XmlHeader {
  return &XmlHeader{
    Type:       xml.readUint16(),
    HeaderSize: xml.readUint16(),
    Size:       xml.readUint32(),
  }
}

func (xml *Xml) parseXmlStrPool() *XmlStrPool {
  s := xml.pos()
  header := xml.parseXmlHeader()
  strCount := xml.readUint32()
  styleCount := xml.readUint32()
  flags := xml.readUint32()
  strStart := xml.readUint32()
  styleStart := xml.readUint32()
  strOffsets := xml.readUint32Array(strCount)
  styleOffsets := xml.readUint32Array(styleCount)

  var strs []string
  if strCount > 0 && styleCount < math.MaxUint32 {
    e := s + header.Size
    if styleCount > 0 && styleCount < math.MaxUint32 {
      e = s + styleStart
    }
    block := xml.slice(xml.pos(), e)
    strs = make([]string, strCount)
    if flags&0x0100 != 0 {
      // UTF-8
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str8(block, strOffsets[i])
      }
    } else {
      // UTF-16
      for i := uint32(0); i < strCount; i++ {
        b := str16Bytes(block, strOffsets[i])
        // 去掉字符之间的空格
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
  xml.Seek(int64(s+header.Size), io.SeekStart)

  return &XmlStrPool{
    XmlHeader:    header,
    StrCount:     strCount,
    StyleCount:   styleCount,
    Flags:        flags,
    StrStart:     strStart,
    StyleStart:   styleStart,
    StrOffsets:   strOffsets,
    StyleOffsets: styleOffsets,
    Strs:         strs,
    Styles:       nil,
  }
}

func (xml *Xml) parseXmlResId() *XmlResId {
  header := xml.parseXmlHeader()
  ids := xml.readUint32Array((header.Size - 8) / 4)
  return &XmlResId{
    XmlHeader: header,
    Ids:       ids,
  }
}

func (xml *Xml) parseXmlNamespace() *XmlNamespace {
  header := xml.parseXmlHeader()
  lineNumber := xml.readUint32()
  res0 := xml.readUint32()
  prefix := xml.readUint32()
  uri := xml.readUint32()
  return &XmlNamespace{
    XmlHeader:  header,
    LineNumber: lineNumber,
    Res0:       res0,
    Prefix:     prefix,
    Uri:        uri,
  }
}

func (xml *Xml) parseXmlStartTag() *XmlTag {
  header := xml.parseXmlHeader()
  lineNumber := xml.readUint32()
  res0 := xml.readUint32()
  uri := xml.readUint32()
  name := xml.readUint32()
  flags := xml.readUint32()
  attrCount := xml.readUint32()
  classAttr := xml.readUint32()

  var attrs []*XmlAttr
  if attrCount > 0 && attrCount < math.MaxUint32 {
    attrs = make([]*XmlAttr, 0, attrCount)
    for i := uint32(0); i < attrCount; i++ {
      attrs = append(attrs, xml.parseXmlAttr())
    }
  }

  return &XmlTag{
    XmlHeader:    header,
    LineNumber:   lineNumber,
    Res0:         res0,
    NamespaceUri: uri,
    Name:         name,
    Flags:        flags,
    AttrCount:    attrCount,
    ClassAttr:    classAttr,
    Attrs:        attrs,
  }
}

func (xml *Xml) parseXmlEndTag() *XmlTag {
  return &XmlTag{
    XmlHeader:    xml.parseXmlHeader(),
    LineNumber:   xml.readUint32(),
    Res0:         xml.readUint32(),
    NamespaceUri: xml.readUint32(),
    Name:         xml.readUint32(),
  }
}

func (xml *Xml) parseXmlAttr() *XmlAttr {
  return &XmlAttr{
    Namespace: xml.readUint32(),
    Uri:       xml.readUint32(),
    Name:      xml.readUint32(),
    Value:     xml.readUint32() >> 24,
    Data:      xml.readUint32(),
  }
}
