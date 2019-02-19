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

// ResStrPool header，28个字节（12-40）
type ResStrPoolHeader struct {
  ResHeader

  // 字符串个数
  StrCount uint32

  // 样式个数
  StyleCount uint32

  // UTF-8或UTF-16
  // SortedFlag = 1<<0
  // UTF8Flag = 1<<8
  Flags uint32

  // 字符串起始位置（相对header）
  StrStart uint32

  // 样式起始位置（相对header）
  StyleStart uint32
}

type ResStrPoolChunk struct {
  Header ResStrPoolHeader

  // 字符串索引，是Strs中每一个字符串的起始位置
  StrIndex []uint32

  // 样式索引，是Styles中每一个样式的起始位置
  StyleIndex []uint32

  // 字符串
  Strs []string

  // 样式
  Styles []string
}

// ResTable header，12个字节（0-12）
type ResTableHeader struct {
  ResHeader

  // package资源包个数，通常一个app只有一个资源包
  PackageCount uint32
}

func main() {
  f, _ := ioutil.ReadFile("C:\\Users\\WangFeng\\Desktop\\resources.arsc")
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
