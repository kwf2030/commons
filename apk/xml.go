package main

import (
  "bytes"
  "fmt"
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

type Xml struct {
  *bytesReader
  MagicNumber uint32
  FileSize    uint32
  StrPool     *XmlStrPool
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
      fmt.Println("UTF8")
      // UTF-8
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str8(block, strOffsets[i])
      }
    } else {
      fmt.Println("UTF16")
      // UTF-16
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str16(block, strOffsets[i])
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
