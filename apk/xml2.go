package apk

import (
  "encoding/json"
  "errors"
  "io/ioutil"
  "math"
  "os"
  "strconv"
  "strings"
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
  ret.NamespacePrefixes = ret.collectNamespacePrefixes()
  ret.Tags2 = ret.collectTags()
  return ret
}

func (xml2 *Xml2) collectNamespacePrefixes() map[uint32]string {
  ret := make(map[uint32]string, 4)
  for _, ns := range xml2.Ori.Namespaces {
    if ns.Prefix < math.MaxUint32 {
      ret[ns.Uri] = xml2.Ori.StrPool.Strs[ns.Prefix]
    }
  }
  return ret
}

func (xml2 *Xml2) collectTags() []*XmlTag2 {
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

func (xml2 *Xml2) addStr(str string) uint32 {
  if str == "" {
    return math.MaxUint32
  }
  pool := xml2.Ori.StrPool
  // 如果已经存在字符串池中，返回其索引
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
  xml2.Ori.Size += 4 + strLen

  return strCount - 1
}

// value只能是数字/字符串/布尔值
func (xml2 *Xml2) AddAttr(key string, value interface{}, f func(*XmlTag2) bool) error {
  attr := &XmlAttr{
    NamespaceUri: math.MaxUint32,
    Name:         math.MaxUint32,
    RawValue:     math.MaxUint32,
    ValueSize:    8,
    DataType:     math.MaxUint8,
    Data:         math.MaxUint32,
  }
  if arr := strings.Split(key, ":"); len(arr) != 2 || arr[0] == "" {
    attr.Name = xml2.addStr(key)
  } else {
    for ns, prefix := range xml2.NamespacePrefixes {
      if arr[0] == prefix {
        attr.NamespaceUri = ns
        break
      }
    }
    if attr.NamespaceUri != math.MaxUint32 {
      attr.Name = xml2.addStr(arr[1])
    } else {
      attr.Name = xml2.addStr(key)
    }
  }
  var valueStr string
  switch val := value.(type) {
  case bool:
    valueStr = "true"
    attr.DataType = 18
    if !val {
      valueStr = "false"
      attr.Data = 0
    }
  case int:
    valueStr = strconv.Itoa(val)
    attr.DataType = 16
    attr.Data = uint32(val)
  case string:
    valueStr = val
    attr.DataType = 3
    attr.Data = xml2.addStr(val)
    attr.RawValue = attr.Data
  default:
    return errors.New("value type invalid")
  }
  var tag2 *XmlTag2
  for _, t := range xml2.Tags2 {
    if f(t) {
      tag2 = t
      break
    }
  }
  if tag2 == nil {
    return errors.New("tag not found")
  }
  tag2.Attrs = append(tag2.Attrs, key+"=\""+valueStr+"\"")
  tag2.Ori.Attrs = append(tag2.Ori.Attrs, attr)
  tag2.Ori.AttrCount += 1
  tag2.Ori.Size += 20
  xml2.Ori.Size += 20
  return nil
}

func (xml2 *Xml2) WriteToFile(name string) error {
  f, e := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
  if e != nil {
    return e
  }
  xml2.Ori.writeTo(newBytesWriter(f))
  f.Close()
  return nil
}

func (xml2 *Xml2) WriteToJsonFile(name string) error {
  data, e := json.Marshal(xml2)
  if e != nil {
    return e
  }
  return ioutil.WriteFile(name, data, os.ModePerm)
}
