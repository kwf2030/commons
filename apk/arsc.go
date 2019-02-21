package main

import (
  "io/ioutil"

  "github.com/kwf2030/commons/conv"
)

type Header struct {
  // chunk类型
  Type uint16

  // chunk header大小
  HeaderSize uint16

  // chunk大小（header + data）
  Size uint32
}

type ResHeader struct {
  Header

  // package资源包个数，通常一个app只有一个资源包
  PackageCount uint32
}

// 28+StrCount*4+StyleCount*4+len(Strs)+len(Styles)
type ResStrPool struct {
  Header

  // 字符串个数
  StrCount uint32

  // 字符串样式个数
  StyleCount uint32

  // 字符串标识，
  // SortedFlag: 0x0001
  // UTF16Flag:  0x0000
  // UTF8Flag:   0x0100
  Flags uint32

  // 字符串起始位置偏移（相对header）
  StrStart uint32

  // 字符串样式起始位置偏移（相对header）
  StyleStart uint32

  // 字符串偏移数组，长度为StrCount
  StrOffsets []uint32

  // 字符串样式偏移数组，长度为StyleCount
  StyleOffsets []uint32

  // 字符串，每个字符串前两个字节为长度，
  // 若是UTF-8编码，以0x00（1个字节）作为结束符，
  // 若是UTF-16编码，以0x0000（2个字节）作为结束符
  Strs []string

  // 字符串样式
  Styles []string
}

type ResPackage struct {
  Header

  // 包Id，用户包是0x7F，系统包是0x01
  Id uint32

  // 包名
  Name string

  // 资源类型字符串池起始位置偏移（相对header）
  TypeStrPoolStart uint32

  // 资源类型个数
  TypeCount uint32

  // 资源项名称字符串池起始位置偏移（相对header）
  EntryStrPoolStart uint32

  // 资源项名称个数
  EntryCount uint32

  // 暂时没用
  TypeIdOffset uint32

  // 资源类型字符串池
  TypeStrPool ResStrPool

  // 资源项名称字符串池
  EntryStrPool ResStrPool
}

type ResTable struct {
  data []byte

  Header        ResHeader
  GlobalStrPool ResStrPool
  Package       ResPackage
}

func ParseResTable(file string) *ResTable {
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return nil
  }
  ret := &ResTable{data: data}
  ret.parseHeader()
  ret.parseGlobalStrPool()
  ret.parsePackage()
  return ret
}

func (r *ResTable) parseHeader() {
  r.Header = ResHeader{
    Header:       parseHeader(r.data, 0),
    PackageCount: conv.BytesToUint32L(r.data[8:12]),
  }
}

func (r *ResTable) parseGlobalStrPool() {
  r.GlobalStrPool = parseStrPool(r.data, 12)
}

func (r *ResTable) parsePackage() {
  p := 12 + r.GlobalStrPool.Size

  header := parseHeader(r.data, p)
  id := conv.BytesToUint32L(r.data[p+8 : p+12])
  // 包名是固定的256个字节，不足的会填充0，
  // UTF-16编码，字符之间也有0，需要去掉
  arr := make([]byte, 0, 128)
  for _, v := range r.data[p+12 : p+268] {
    if v != 0 {
      arr = append(arr, v)
    }
  }
  name := string(arr)
  typeStrPoolStart := conv.BytesToUint32L(r.data[p+268 : p+272])
  typeCount := conv.BytesToUint32L(r.data[p+272 : p+276])
  entryStrPoolStart := conv.BytesToUint32L(r.data[p+276 : p+280])
  entryCount := conv.BytesToUint32L(r.data[p+280 : p+284])
  typeIdOffset := conv.BytesToUint32L(r.data[p+284 : p+288])

  r.Package = ResPackage{
    Header:            header,
    Id:                id,
    Name:              name,
    TypeStrPoolStart:  typeStrPoolStart,
    TypeCount:         typeCount,
    EntryStrPoolStart: entryStrPoolStart,
    EntryCount:        entryCount,
    TypeStrPool:       parseStrPool(r.data, p+typeStrPoolStart),
    EntryStrPool:      parseStrPool(r.data, p+entryStrPoolStart),
    TypeIdOffset:      typeIdOffset,
  }
}

func parseHeader(data []byte, offset uint32) Header {
  return Header{
    Type:       conv.BytesToUint16L(data[offset : offset+2]),
    HeaderSize: conv.BytesToUint16L(data[offset+2 : offset+4]),
    Size:       conv.BytesToUint32L(data[offset+4 : offset+8]),
  }
}

func parseStrPool(data []byte, offset uint32) ResStrPool {
  header := parseHeader(data, offset)
  strCount := conv.BytesToUint32L(data[offset+8 : offset+12])
  styleCount := conv.BytesToUint32L(data[offset+12 : offset+18])
  flags := conv.BytesToUint32L(data[offset+16 : offset+20])
  strStart := conv.BytesToUint32L(data[offset+20 : offset+24])
  styleStart := conv.BytesToUint32L(data[offset+24 : offset+28])

  var strOffsets []uint32
  if strCount > 0 {
    strOffsets = make([]uint32, strCount)
    var s, e uint32
    for i := uint32(0); i < strCount; i++ {
      s = offset + 28 + i*4
      e = s + 4
      strOffsets[i] = conv.BytesToUint32L(data[s:e])
    }
  }

  var styleOffsets []uint32
  if styleCount > 0 {
    styleOffsets = make([]uint32, styleCount)
    var s, e uint32
    for i := uint32(0); i < styleCount; i++ {
      s = offset + 28 + strCount*4 + i*4
      e = s + 4
      styleOffsets[i] = conv.BytesToUint32L(data[s:e])
    }
  }

  var strs []string
  if strCount > 0 {
    strs = make([]string, strCount)
    s := offset + strStart
    var e uint32
    if styleCount > 0 {
      e = offset + styleStart
    } else {
      e = offset + header.Size
    }
    arr := data[s:e]
    if flags&0x0100 != 0 {
      // UTF-8
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str8(arr, strOffsets[i])
      }
    } else {
      // UTF-16
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str16(arr, strOffsets[i])
      }
    }
  }

  // todo parse style strings
  return ResStrPool{
    Header:       header,
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
