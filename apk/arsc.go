package main

import (
  "io/ioutil"

  "github.com/kwf2030/commons/conv"
)

// Size: 8
type Header struct {
  // chunk类型
  Type uint16

  // chunk header大小
  HeaderSize uint16

  // chunk大小（包括header）
  Size uint32
}

func parseHeader(data []byte, offset uint32) Header {
  return Header{
    Type:       conv.BytesToUint16L(data[offset : offset+2]),
    HeaderSize: conv.BytesToUint16L(data[offset+2 : offset+4]),
    Size:       conv.BytesToUint32L(data[offset+4 : offset+8]),
  }
}

// Size: 12
type ResHeader struct {
  Header

  // 资源包个数，通常一个app只有一个资源包
  PackageCount uint32
}

// Header Size: 28
// Size: 28+StrCount*4+StyleCount*4+Strs+Styles
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

func str8(data []byte, offset uint32) string {
  n := 1
  if x := data[offset] & 0x80; x != 0 {
    n = 2
  }
  s := offset + uint32(n)
  l := data[s]
  if l == 0 {
    return ""
  }
  s++
  if l&0x80 != 0 {
    l = (l&0x7F)<<8 | data[s]&0xFF
    s++
  }
  return string(data[s : s+uint32(l)])
}

func str16(data []byte, offset uint32) string {
  n := 2
  if x := data[offset+1] & 0x80; x != 0 {
    n = 4
  }
  s := offset + uint32(n)
  e := s
  l := uint32(len(data))
  for {
    if e+1 >= l {
      break
    }
    if data[e] == 0 && data[e+1] == 0 {
      break
    }
    e += 2
  }
  return string(data[s:e])
}

// Header Size: 288
// Size: 288+TypeStrPool+KeyStrPool+TypeSpecs
type ResPackage struct {
  Header

  // 包Id，用户包是0x7F，系统包是0x01
  Id uint32

  // 包名（原本256个字节，这里已经把多余的字节去掉了）
  Name string

  // 资源类型字符串池起始位置偏移（相对header）
  TypeStrPoolStart uint32

  // 资源类型个数
  TypeCount uint32

  // 资源项名称字符串池起始位置偏移（相对header）
  KeyStrPoolStart uint32

  // 资源项名称个数
  KeyCount uint32

  // 保留字段
  Res0 uint32

  // 资源类型字符串池
  TypeStrPool ResStrPool

  // 资源项名称字符串池
  KeyStrPool ResStrPool

  Types     []ResType
  TypeSpecs []ResTypeSpec
}

func parsePackage(data []byte, offset uint32) ResPackage {
  header := parseHeader(data, offset)
  id := conv.BytesToUint32L(data[offset+8 : offset+12])
  // 包名是固定的256个字节，不足的会填充0，
  // UTF-16编码，每2个字节表示一个字符，所以字符之间会有0，需要去掉
  arr := make([]byte, 0, 128)
  for _, v := range data[offset+12 : offset+268] {
    if v != 0 {
      arr = append(arr, v)
    }
  }
  name := string(arr)
  typeStrPoolStart := conv.BytesToUint32L(data[offset+268 : offset+272])
  typeCount := conv.BytesToUint32L(data[offset+272 : offset+276])
  keyStrPoolStart := conv.BytesToUint32L(data[offset+276 : offset+280])
  keyCount := conv.BytesToUint32L(data[offset+280 : offset+284])
  res0 := conv.BytesToUint32L(data[offset+284 : offset+288])
  typeStrPool := parseStrPool(data, offset+typeStrPoolStart)
  keyStrPool := parseStrPool(data, offset+keyStrPoolStart)

  var typeSpecs []ResTypeSpec
  var types []ResType
  if typeCount > 0 {
    typeSpecs = make([]ResTypeSpec, 0, typeCount)
    types = make([]ResType, 0, typeCount)
    l := uint32(len(data))
    offset += keyStrPoolStart + keyStrPool.Size
    for offset < l {
      switch conv.BytesToUint16L(data[offset : offset+2]) {
      case 0x0201:
        // Type
        t := parseType(data, offset)
        types = append(types, t)
        offset += t.Size
      case 0x0202:
        // Type Spec
        t := parseTypeSpec(data, offset)
        typeSpecs = append(typeSpecs, t)
        offset += t.Size
      default:
        offset += 2
      }
    }
  }

  return ResPackage{
    Header:           header,
    Id:               id,
    Name:             name,
    TypeStrPoolStart: typeStrPoolStart,
    TypeCount:        typeCount,
    KeyStrPoolStart:  keyStrPoolStart,
    KeyCount:         keyCount,
    Res0:             res0,
    TypeStrPool:      typeStrPool,
    KeyStrPool:       keyStrPool,
    Types:            types,
    TypeSpecs:        typeSpecs,
  }
}

// Header Size: 16
// Size: 16+EntryCount*4
type ResTypeSpec struct {
  Header

  // 资源类型Id
  Id uint8

  // 两个保留字段
  Res0 uint8
  Res1 uint16

  // 资源项个数
  EntryCount uint32

  // 资源项标记，长度为EntryCount
  EntryFlags []uint32
}

func parseTypeSpec(data []byte, offset uint32) ResTypeSpec {
  header := parseHeader(data, offset)
  id := uint8(data[offset+8])
  res0 := uint8(data[offset+9])
  res1 := conv.BytesToUint16L(data[offset+10 : offset+12])
  entryCount := conv.BytesToUint32L(data[offset+12 : offset+16])

  var entryFlags []uint32
  if entryCount > 0 {
    entryFlags = make([]uint32, entryCount)
    var s, e uint32
    for i := uint32(0); i < entryCount; i++ {
      s = offset + 16 + i*4
      e = s + 4
      entryFlags[i] = conv.BytesToUint32L(data[s:e])
    }
  }

  return ResTypeSpec{
    Header:     header,
    Id:         id,
    Res0:       res0,
    Res1:       res1,
    EntryCount: entryCount,
    EntryFlags: entryFlags,
  }
}

// Header Size: 76
// Size: 76+EntryCount*4+Entries
type ResType struct {
  Header

  // 资源类型Id
  Id uint8

  // 两个保留字段
  Res0 uint8
  Res1 uint16

  // 资源项个数
  EntryCount uint32

  // 资源项起始位置偏移（相对header）
  EntryStart uint32

  // 配置描述
  Config ResConfig

  // 资源项偏移数组，长度为EntryCount
  EntryOffsets []uint32

  // 资源项
  Entries []ResEntry
}

func parseType(data []byte, offset uint32) ResType {
  header := parseHeader(data, offset)
  id := uint8(data[offset+8])
  res0 := uint8(data[offset+9])
  res1 := conv.BytesToUint16L(data[offset+10 : offset+12])
  entryCount := conv.BytesToUint32L(data[offset+12 : offset+16])
  entryStart := conv.BytesToUint32L(data[offset+16 : offset+20])
  config := parseConfig(data, offset+20)

  var entryOffsets []uint32
  if entryCount > 0 {
    entryOffsets = make([]uint32, entryCount)
    var s, e uint32
    for i := uint32(0); i < entryCount; i++ {
      s = offset + 76 + i*4
      e = s + 4
      // 可能存在无效偏移值（math.MaxUint32）
      entryOffsets[i] = conv.BytesToUint32L(data[s:e])
    }
  }

  return ResType{
    Header:       header,
    Id:           id,
    Res0:         res0,
    Res1:         res1,
    EntryCount:   entryCount,
    EntryStart:   entryStart,
    Config:       config,
    EntryOffsets: entryOffsets,
  }
}

// Size: 56
type ResConfig struct {
  Size                  uint32
  Mcc                   uint16
  Mnc                   uint16
  Language              uint16
  Country               uint16
  Orientation           uint8
  Touchscreen           uint8
  Density               uint16
  Keyboard              uint8
  Navigation            uint8
  InputFlags            uint8
  InputPad0             uint8
  ScreenWidth           uint16
  ScreenHeight          uint16
  SdkVersion            uint16
  MinorVersion          uint16
  ScreenLayout          uint8
  UiMode                uint8
  SmallestScreenWidthDp uint16
  ScreenWidthDp         uint16
  ScreenHeightDp        uint16
}

func parseConfig(data []byte, offset uint32) ResConfig {
  return ResConfig{
    Size:                  conv.BytesToUint32L(data[offset : offset+4]),
    Mcc:                   conv.BytesToUint16L(data[offset+4 : offset+6]),
    Mnc:                   conv.BytesToUint16L(data[offset+6 : offset+8]),
    Language:              conv.BytesToUint16L(data[offset+8 : offset+10]),
    Country:               conv.BytesToUint16L(data[offset+10 : offset+12]),
    Orientation:           uint8(data[offset+12]),
    Touchscreen:           uint8(data[offset+13]),
    Density:               conv.BytesToUint16L(data[offset+14 : offset+16]),
    Keyboard:              uint8(data[offset+16]),
    Navigation:            uint8(data[offset+17]),
    InputFlags:            uint8(data[offset+18]),
    InputPad0:             uint8(data[offset+19]),
    ScreenWidth:           conv.BytesToUint16L(data[offset+20 : offset+22]),
    ScreenHeight:          conv.BytesToUint16L(data[offset+22 : offset+24]),
    SdkVersion:            conv.BytesToUint16L(data[offset+24 : offset+26]),
    MinorVersion:          conv.BytesToUint16L(data[offset+26 : offset+28]),
    ScreenLayout:          uint8(data[offset+28]),
    UiMode:                uint8(data[offset+29]),
    SmallestScreenWidthDp: conv.BytesToUint16L(data[offset+30 : offset+32]),
    ScreenWidthDp:         conv.BytesToUint16L(data[offset+32 : offset+34]),
    ScreenHeightDp:        conv.BytesToUint16L(data[offset+34 : offset+36]),
  }
}

type ResEntry struct {
}

func parseEntry(data []byte, offset uint32) ResEntry {
  return ResEntry{}
}

type ResMapEntry struct {
}

func parseMapEntry(data []byte, offset uint32) ResMapEntry {
  return ResMapEntry{}
}

type ResValue struct {
}

func parseValue(data []byte, offset uint32) ResValue {
  return ResValue{}
}

// Header Size: 12
// Size: resources.arsc
type ResTable struct {
  ResHeader
  StrPool  ResStrPool
  Packages []ResPackage
}

func ParseResTable(file string) *ResTable {
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return nil
  }
  header := ResHeader{
    Header:       parseHeader(data, 0),
    PackageCount: conv.BytesToUint32L(data[8:12]),
  }
  strPool := parseStrPool(data, 12)
  packages := make([]ResPackage, header.PackageCount)
  offset := 12 + strPool.Size
  for i := uint32(0); i < header.PackageCount; i++ {
    packages[i] = parsePackage(data, offset)
    offset += packages[i].Size
  }
  return &ResTable{
    ResHeader: header,
    StrPool:   strPool,
    Packages:  packages,
  }
}
