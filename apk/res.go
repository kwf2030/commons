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
  // 第0-7个字节
  ResHeader

  // package资源包个数，通常一个app只有一个资源包，
  // 第8-11个字节
  PackageCount uint32
}

// 28个字节
type ResStrPoolHeader struct {
  // 第12-19个字节
  ResHeader

  // 字符串个数，
  // 第20-23个字节
  StrCount uint32

  // 字符串样式个数，
  // 第24-27个字节
  StyleCount uint32

  // 字符串标识，
  // SortedFlag = 1
  // UTF16Flag = 0
  // UTF8Flag = 1<<8(0x0100)
  // 第28-31个字节
  Flags uint32

  // 字符串起始位置偏移（相对header），
  // 第32-35个字节
  StrStart uint32

  // 字符串样式起始位置偏移（相对header），
  // 第36-39个字节
  StyleStart uint32
}

// 28+StrCount*4+StyleCount*4+字符串大小+字符串样式大小
type ResStrPoolChunk struct {
  // 第12-39个字节
  Header ResStrPoolHeader

  // 字符串偏移数组，其元素对应Strs中每一个元素的起始位置，
  // 长度为Header.StrCount，
  // 第40-[(40+Header.StrCount*4)-1]个字节
  StrOffsets []uint32

  // 字符串样式偏移数组，其元素对应Styles中每一个元素的起始位置，
  // 长度为Header.StyleCount,
  // 第(40+Header.StrCount*4)-[(40+Header.StrCount*4)+(Header.StyleCount*4)-1]
  StyleOffsets []uint32

  // 字符串，前两个字节为长度，计算方法为：(((hbyte & 0x7F) << 8)) | lbyte
  // 若是UTF-8编码，以0x00作为结束符，
  // 若是UTF-16编码，以0x0000作为结束符，
  // 与Styles一一对应，即如果第N个字符串有样式，那么其样式应该是Styles的第N个元素，
  // 第(28+Header.StrStart)-[(28+Header.StyleStart)-1]个字节
  Strs []string

  // 字符串样式，
  // 第(28+Header.StyleStart)-[(28+Header.Size)-1]
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
  f, _ := ioutil.ReadFile("/home/wangfeng/workspace/wechat/tmp/resources.arsc")
  // t := conv.BytesToUint16L(f[12:14])
  // h := conv.BytesToUint16L(f[14:16])
  // s := conv.BytesToUint32L(f[16:20])
  c1 := conv.BytesToUint32L(f[20:24])
  c2 := conv.BytesToUint32L(f[24:28])
  flags := conv.BytesToUint32L(f[28:32])
  ss1 := conv.BytesToUint32L(f[32:36])
  ss2 := conv.BytesToUint32L(f[36:40])
  fmt.Println(c1, c2, flags, ss1, ss2)
}
