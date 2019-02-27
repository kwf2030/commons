package main

import (
  "bytes"
  "io"
  "io/ioutil"
  "math"
)

type XmlStrPool struct {
  Type         uint32
  Size         uint32
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
  Type uint32
  Size uint32
  Ids  []uint32
}

type XmlNamespace struct {
  Type       uint32
  Size       uint32
  LineNumber uint32
  Res0       uint32
  Prefix     uint32
  Uri        uint32
}

type XmlTag struct {
  Type         uint32
  Size         uint32
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
  MagicNumber uint32
  FileSize    uint32
  StrPool     *XmlStrPool
  ResId       *XmlResId
  Namespace   *XmlNamespace
  Tag         *XmlTag
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
  xml.MagicNumber = xml.readUint32()
  xml.FileSize = xml.readUint32()
  xml.StrPool = xml.parseXmlStrPool()
  xml.ResId = xml.parseXmlResId()
  xml.Namespace = xml.parseXmlNamespace()
  xml.Tag = xml.parseXmlTag()
  return xml
}

func (xml *Xml) parseXmlStrPool() *XmlStrPool {
  s := xml.pos()
  tp := xml.readUint32()
  size := xml.readUint32()
  strCount := xml.readUint32()
  styleCount := xml.readUint32()
  flags := xml.readUint32()
  strStart := xml.readUint32()
  styleStart := xml.readUint32()
  strOffsets := xml.readUint32Array(strCount)
  styleOffsets := xml.readUint32Array(styleCount)

  var strs []string
  if strCount > 0 && styleCount < math.MaxUint32 {
    e := s + size
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
  xml.Seek(int64(s+size), io.SeekStart)

  return &XmlStrPool{
    Type:         tp,
    Size:         size,
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
  tp := xml.readUint32()
  size := xml.readUint32()
  ids := xml.readUint32Array((size - 8) / 4)
  return &XmlResId{
    Type: tp,
    Size: size,
    Ids:  ids,
  }
}

func (xml *Xml) parseXmlNamespace() *XmlNamespace {
  tp := xml.readUint32()
  size := xml.readUint32()
  lineNumber := xml.readUint32()
  res0 := xml.readUint32()
  prefix := xml.readUint32()
  uri := xml.readUint32()
  return &XmlNamespace{
    Type:       tp,
    Size:       size,
    LineNumber: lineNumber,
    Res0:       res0,
    Prefix:     prefix,
    Uri:        uri,
  }
}

func (xml *Xml) parseXmlTag() *XmlTag {
  tp := xml.readUint32()
  size := xml.readUint32()
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
    Type:         tp,
    Size:         size,
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

func (xml *Xml) parseXmlAttr() *XmlAttr {
  return &XmlAttr{
    Namespace: xml.readUint32(),
    Uri:       xml.readUint32(),
    Name:      xml.readUint32(),
    Value:     xml.readUint32(),
    Data:      xml.readUint32(),
  }
}
