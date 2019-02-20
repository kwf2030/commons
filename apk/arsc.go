package main

import (
  "io/ioutil"

  "github.com/kwf2030/commons/conv"
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
  // SortedFlag = 0x0001
  // UTF16Flag = 0x0000
  // UTF8Flag = 0x0100
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
type ResStrPool struct {
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

  // 字符串，前两个字节为长度，
  // 若是UTF-8编码，以0x00（1个字节）作为结束符，
  // 若是UTF-16编码，以0x0000（2个字节）作为结束符，
  // 起始：12+Header.StrStart，
  // 结束：
  Strs []string

  // 字符串样式，
  // 起始：12+Header.StyleStart，
  // 结束：12+Header.Size
  Styles []string
}

type ResTablePackage struct {
  Header ResHeader

  // 包Id，用户包Id是0x7F，系统包Id是0x01
  Id uint32

  // 包名
  Name string

  TypeStrs       uint32
  LasrPublicType uint32
  KeyStrs        uint32
  LastPublicKey  uint32
}

type ResArsc struct {
  data []byte

  TableHeader   ResTableHeader
  GlobalStrPool ResStrPool
}

func ParseResArsc(file string) *ResArsc {
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return nil
  }
  ret := &ResArsc{data: data}
  ret.parseTableHeader()
  ret.parseGlobalStrPool()
  return ret
}

func (r *ResArsc) parseTableHeader() {
  r.TableHeader = ResTableHeader{
    ResHeader:    ResHeader{conv.BytesToUint16L(r.data[:2]), conv.BytesToUint16L(r.data[2:4]), conv.BytesToUint32L(r.data[4:8])},
    PackageCount: conv.BytesToUint32L(r.data[8:12]),
  }
}

func (r *ResArsc) parseGlobalStrPool() {
  header := ResStrPoolHeader{
    ResHeader:  ResHeader{conv.BytesToUint16L(r.data[12:14]), conv.BytesToUint16L(r.data[14:16]), conv.BytesToUint32L(r.data[16:20])},
    StrCount:   conv.BytesToUint32L(r.data[20:24]),
    StyleCount: conv.BytesToUint32L(r.data[24:28]),
    Flags:      conv.BytesToUint32L(r.data[28:32]),
    StrStart:   conv.BytesToUint32L(r.data[32:36]),
    StyleStart: conv.BytesToUint32L(r.data[36:40]),
  }

  var strOffsets []uint32
  if header.StrCount > 0 {
    strOffsets = make([]uint32, header.StrCount)
    var s, e uint32
    for i := uint32(0); i < header.StrCount; i++ {
      s = 40 + i*4
      e = s + 4
      strOffsets[i] = conv.BytesToUint32L(r.data[s:e])
    }
  }

  var styleOffsets []uint32
  if header.StyleCount > 0 {
    styleOffsets = make([]uint32, header.StyleCount)
    var s, e uint32
    for i := uint32(0); i < header.StyleCount; i++ {
      s = 40 + i*4
      e = s + 4
      styleOffsets[i] = conv.BytesToUint32L(r.data[s:e])
    }
  }

  var strs []string
  if header.StrCount > 0 {
    strs = make([]string, header.StrCount)
    s := 12 + header.StrStart
    var e uint32
    if header.StyleCount > 0 {
      e = 12 + header.StyleStart
    } else {
      e = 12 + header.Size
    }
    arr := r.data[s:e]
    if header.Flags&0x0100 != 0 {
      // UTF-8
      for i := uint32(0); i < header.StrCount; i++ {
        strs[i] = str8(arr, strOffsets[i])
      }
    } else {
      // UTF-16
      for i := uint32(0); i < header.StrCount; i++ {
        strs[i] = str16(arr, strOffsets[i])
      }
    }
  }

  // todo parse style strings
  r.GlobalStrPool = ResStrPool{
    Header:       header,
    StrOffsets:   strOffsets,
    StyleOffsets: styleOffsets,
    Strs:         strs,
    Styles:       nil,
  }
}

func str8(arr []byte, offset uint32) string {
  n := 1
  if x := arr[offset] & 0x80; x != 0 {
    n = 2
  }
  s := offset + uint32(n)
  l := arr[s]
  if l == 0 {
    return ""
  }
  s++
  if l&0x80 != 0 {
    l = (l&0x7F)<<8 | arr[s]&0xFF
    s++
  }
  return string(arr[s : s+uint32(l)])
}

func str16(arr []byte, offset uint32) string {
  n := 2
  if x := arr[offset+1] & 0x80; x != 0 {
    n = 4
  }
  s := offset + uint32(n)
  e := s
  l := uint32(len(arr))
  for {
    if e+1 >= l {
      break
    }
    if arr[e] == 0 && arr[e+1] == 0 {
      break
    }
    e += 2
  }
  return string(arr[s:e])
}
