package main

import (
  "math"
  "strconv"
)

type XmlTag2 struct {
  Name  string
  Attrs []string
}

type Xml2 struct {
  *Xml
  NamespacePrefixes map[uint32]string
  Tags2             []*XmlTag2
}

func NewXml2(xml *Xml) *Xml2 {
  ret := &Xml2{Xml: xml}
  ret.NamespacePrefixes = ret.CollectNamespacePrefixes()
  ret.Tags2 = ret.CollectTags()
  return ret
}

func (xml2 *Xml2) CollectNamespacePrefixes() map[uint32]string {
  ret := make(map[uint32]string, 4)
  for _, ns := range xml2.Namespaces {
    if ns.Prefix < math.MaxUint32 {
      ret[ns.Uri] = xml2.StrPool.Strs[ns.Prefix]
    }
  }
  return ret
}

func (xml2 *Xml2) CollectTags() []*XmlTag2 {
  ret := make([]*XmlTag2, 0, len(xml2.Tags))
  for _, tag := range xml2.Tags {
    tagName := xml2.StrPool.Strs[tag.Name]
    if tag.NamespaceUri < math.MaxUint32 {
      tagName = xml2.NamespacePrefixes[tag.NamespaceUri] + ":" + tagName
    }
    if tag.AttrCount <= 0 {
      ret = append(ret, &XmlTag2{Name: tagName})
      continue
    }
    attrs := make([]string, tag.AttrCount)
    for i, attr := range tag.Attrs {
      attrName := xml2.StrPool.Strs[attr.Name]
      if attr.NamespaceUri < math.MaxUint32 {
        attrName = xml2.NamespacePrefixes[attr.NamespaceUri] + ":" + attrName
      }
      attrVal := xml2.parseData(attr.DataType, attr.Data)
      attrs[i] = attrName + "=" + attrVal
    }
    ret = append(ret, &XmlTag2{Name: tagName, Attrs: attrs})
  }
  return ret
}

func (xml2 *Xml2) parseData(dataType uint8, data uint32) string {
  switch dataType {
  case 3:
    if data < math.MaxUint32 {
      return xml2.StrPool.Strs[data]
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
