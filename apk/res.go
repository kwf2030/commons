package main

import (
  "fmt"
  "io/ioutil"

  "github.com/kwf2030/commons/conv"
)

// chunk类型
const (
  ResNull    = 0x0000
  ResStrPool = 0x0001
  ResTable   = 0x0002
  ResXml     = 0x0003
)

// ResTable子类型
const (
  ResTablePackage     = 0x0200
  ResTableType        = 0x0201
  ResTableTypeSpec    = 0x0202
  ResTableTypeLibrary = 0x0203
)

// ResXml子类型
const (
  ResXmlFirstChunk     = 0x0100
  ResXmlStartNamespace = 0x0100
  ResXmlEndNamespace   = 0x0101
  ResXmlStartElement   = 0x0102
  ResXmlEndElement     = 0x0103
  ResXmlCData          = 0x0104
  ResXmlLastChunk      = 0x017F
  ResXmlResourceMap    = 0x0180
)

// 8个字节
type ResHeader struct {
  // chunk类型
  Type uint16

  // chunk header大小
  HeaderSize uint16

  // chunk大小（header + data）
  Size uint32
}

// 12个字节
type ResTableHeader struct {
  // 起始：0，
  // 结束：8
  ResHeader

  // package资源包个数，通常一个app只有一个资源包，
  // 起始：8，
  // 结束：12
  PackageCount uint32
}

// 28个字节
type ResStrPoolHeader struct {
  // 起始：12，
  // 结束：20
  ResHeader

  // 字符串个数，
  // 起始：20，
  // 结束：24
  StrCount uint32

  // 字符串样式个数，
  // 起始：24，
  // 结束：28
  StyleCount uint32

  // 字符串标识，
  // SortedFlag = 1
  // UTF16Flag = 0
  // UTF8Flag = 1<<8(0x0100)
  // 起始：28，
  // 结束：32
  Flags uint32

  // 字符串起始位置偏移（相对header），
  // 起始：32，
  // 结束：36
  StrStart uint32

  // 字符串样式起始位置偏移（相对header），
  // 起始：36，
  // 结束：40
  StyleStart uint32
}

// 28+StrCount*4+StyleCount*4+字符串大小+字符串样式大小
type ResStrPoolChunk struct {
  // 起始：12，
  // 结束：40
  Header ResStrPoolHeader

  // 字符串偏移数组，其元素对应Strs中每一个元素的起始位置，
  // 长度为Header.StrCount，
  // 起始：40，
  // 结束：40+Header.StrCount*4
  StrOffsets []uint32

  // 字符串样式偏移数组，其元素对应Styles中每一个元素的起始位置，
  // 长度为Header.StyleCount,
  // 起始：40+Header.StrCount*4，
  // 结束：40+Header.StrCount*4+Header.StyleCount*4
  StyleOffsets []uint32

  // 字符串，前两个字节为长度：(((hbyte & 0x7F) << 8)) | lbyte
  // 若是UTF-8编码，以0x00（1个字节）作为结束符，
  // 若是UTF-16编码，以0x0000（2个字节）作为结束符，
  // 与Styles一一对应，即如果Strs[n]有样式，那么其样式是Styles[n]，
  // 起始：12+Header.StrStart，即StyleOffsets的结束位置，
  // 结束：12+Header.StrStart+Header.StrCount*4，即12+Header.StyleStart
  Strs []string

  // 字符串样式，
  // 起始：12+Header.StyleStart，即Strs的结束位置，
  // 结束：12+Header.StyleStart+StyleCount*4，即12+Header.Size
  Styles []string
}

type ResStrRef struct {
  Index uint32
}

type ResStrSpan struct {
  // 样式字符串在字符串池中的偏移
  Name ResStrRef

  // 应用样式的第一个字符
  FirstChar uint32

  // 应用样式的最后一个字符
  LastChar uint32
}

func main() {
  // f, _ := ioutil.ReadFile("/home/wangfeng/workspace/wechat/tmp/resources.arsc")
  f, _ := ioutil.ReadFile("C:\\Users\\WangFeng\\Desktop\\resources.arsc")

  tableHeader := ResTableHeader{
    ResHeader:    ResHeader{conv.BytesToUint16L(f[:2]), conv.BytesToUint16L(f[2:4]), conv.BytesToUint32L(f[4:8])},
    PackageCount: conv.BytesToUint32L(f[8:12]),
  }

  strPoolHeader := ResStrPoolHeader{
    ResHeader:  ResHeader{conv.BytesToUint16L(f[12:14]), conv.BytesToUint16L(f[14:16]), conv.BytesToUint32L(f[16:20])},
    StrCount:   conv.BytesToUint32L(f[20:24]),
    StyleCount: conv.BytesToUint32L(f[24:28]),
    Flags:      conv.BytesToUint32L(f[28:32]),
    StrStart:   conv.BytesToUint32L(f[32:36]),
    StyleStart: conv.BytesToUint32L(f[36:40]),
  }

  fmt.Printf("tableHeader:%+v\n", tableHeader)
  fmt.Printf("strPoolHeader:%+v\n", strPoolHeader)

  var strOffsets []uint32
  if strPoolHeader.StrCount > 0 {
    strOffsets = make([]uint32, strPoolHeader.StrCount)
    var s, e uint32
    i := uint32(0)
    for ; i < strPoolHeader.StrCount; i++ {
      s = 40 + i*4
      e = s + 4
      strOffsets[i] = conv.BytesToUint32L(f[s:e])
    }
  }

  var styleOffsets []uint32
  if strPoolHeader.StyleCount > 0 {
    styleOffsets = make([]uint32, strPoolHeader.StyleCount)
    var s, e uint32
    i := uint32(0)
    for ; i < strPoolHeader.StyleCount; i++ {
      s = 40 + i*4
      e = s + 4
      styleOffsets[i] = conv.BytesToUint32L(f[s:e])
    }
  }

  var strs []string
  if strPoolHeader.StrCount > 0 {
    strs = make([]string, strPoolHeader.StrCount)
    i := uint32(0)
    p := 12 + strPoolHeader.StrStart
    for ; i < strPoolHeader.StrCount; i++ {
      s := p + strOffsets[i] + 2
      if i < strPoolHeader.StrCount-1 {
        strs[i] = string(f[s : p+strOffsets[i+1]])
      } else {
        if strPoolHeader.StyleCount > 0 {
          strs[i] = string(f[s : 12+strPoolHeader.StyleStart])
        } else {
          strs[i] = string(f[s : 12+strPoolHeader.Size])
        }
      }
    }
  }

  strPoolChunk := ResStrPoolChunk{
    Header:       strPoolHeader,
    StrOffsets:   strOffsets,
    StyleOffsets: styleOffsets,
    Strs:         strs,
    Styles:       nil,
  }

  for i := 100; i < 250; i++ {
    fmt.Println(strPoolChunk.Strs[i])
  }
}
