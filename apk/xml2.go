package main

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
  return ret
}

func (xml2 *Xml2) CollectNamespacePrefixes() map[uint32]string {
  return nil
}

func (xml2 *Xml2) CollectTags() []*XmlTag2 {
  return nil
}
