package main

import (
  "bytes"
  "github.com/kwf2030/commons/conv"
  "math"
  "strconv"
)

type XmlTag2 struct {
  Ori   *XmlTag
  Name  string
  Attrs []string
}

type Xml2 struct {
  Ori               *Xml
  NamespacePrefixes map[uint32]string
  Tags2             []*XmlTag2
}

func NewXml2(xml *Xml) *Xml2 {
  ret := &Xml2{Ori: xml}
  ret.NamespacePrefixes = ret.CollectNamespacePrefixes()
  ret.Tags2 = ret.CollectTags()
  return ret
}

func (xml2 *Xml2) CollectNamespacePrefixes() map[uint32]string {
  ret := make(map[uint32]string, 4)
  for _, ns := range xml2.Ori.Namespaces {
    if ns.Prefix < math.MaxUint32 {
      ret[ns.Uri] = xml2.Ori.StrPool.Strs[ns.Prefix]
    }
  }
  return ret
}

func (xml2 *Xml2) CollectTags() []*XmlTag2 {
  ret := make([]*XmlTag2, 0, len(xml2.Ori.Tags))
  for _, tag := range xml2.Ori.Tags {
    tagName := xml2.Ori.StrPool.Strs[tag.Name]
    if tag.NamespaceUri < math.MaxUint32 {
      tagName = xml2.NamespacePrefixes[tag.NamespaceUri] + ":" + tagName
    }
    if tag.AttrCount <= 0 {
      ret = append(ret, &XmlTag2{Ori: tag, Name: tagName})
      continue
    }
    attrs := make([]string, tag.AttrCount)
    for i, attr := range tag.Attrs {
      attrName := xml2.Ori.StrPool.Strs[attr.Name]
      if attr.NamespaceUri < math.MaxUint32 {
        attrName = xml2.NamespacePrefixes[attr.NamespaceUri] + ":" + attrName
      }
      attrVal := xml2.parseData(attr.DataType, attr.Data)
      attrs[i] = attrName + "=\"" + attrVal + "\""
    }
    ret = append(ret, &XmlTag2{Ori: tag, Name: tagName, Attrs: attrs})
  }
  return ret
}

func (xml2 *Xml2) parseData(dataType uint8, data uint32) string {
  switch dataType {
  case 3:
    if data < math.MaxUint32 {
      return xml2.Ori.StrPool.Strs[data]
    }
  case 16:
    return strconv.Itoa(int(data))
  case 18:
    if data == 0 {
      return "false"
    }
    return "true"
  }
  return ""
}

func AddToStrPool(str string, pool XmlStrPool, poolBytes []byte) []byte {
  buf := bytes.Buffer{}
  // Type/HeaderSize
  buf.Write(poolBytes[:4])
  // Size, Size+4(StrOffsets)+str(str+3)
  buf.Write(conv.Uint32ToBytesL(pool.Size + 4 + uint32(len(str)) + 3))
  // StrCount
  buf.Write(conv.Uint32ToBytesL(pool.StrCount + 1))
  // StyleCount/Flags/StrStart
  buf.Write(poolBytes[12:24])
  // StyleStart, StyleStart+4(StrOffsets)+str(str+3)
  if pool.StyleCount > 0 {
    buf.Write(conv.Uint32ToBytesL(pool.StyleStart + 4 + uint32(len(str)) + 3))
  }
  //buf.Write() // StrOffsets, +4
  //buf.Write() // StyleOffsets
  //buf.Write() // Strs
  //buf.Write() // Styles
  return buf.Bytes()
}
