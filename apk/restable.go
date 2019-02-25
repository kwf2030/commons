package main

import (
  "bytes"
  "fmt"
  "io"
  "io/ioutil"

  "github.com/kwf2030/commons/conv"
)

type Header struct {
  // 类型
  Type uint16

  // header大小
  HeaderSize uint16

  // chunk大小（包括header）
  Size uint32
}

type ResStrPool struct {
  *Header

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

  Styles []string
}

type ResPackage struct {
  *Header

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
  TypeStrPool *ResStrPool

  // 资源项名称字符串池
  KeyStrPool *ResStrPool

  TypeSpecs []*ResTypeSpec

  Types []*ResType
}

type ResTypeSpec struct {
  *Header

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

type ResType struct {
  *Header

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
  Config *ResConfig

  // 资源项偏移数组，长度为EntryCount
  EntryOffsets []uint32

  // 资源项
  Entries []*ResEntry
}

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

type ResEntry struct {
  Size      uint16
  Flags     uint16
  Key       uint32
  Value     *ResValue
  ParentRef uint32
  Count     uint32
  Values    []*ResValue
}

type ResValue struct {
  Size     uint16
  Res0     uint8
  DataType uint8
  Data     uint32
}

type ResTable struct {
  *bytesReader

  *Header

  // 资源包个数，通常一个app只有一个资源包
  PackageCount uint32

  // 全局字符串池
  StrPool *ResStrPool

  // 资源包，
  // len(Packages)=PackageCount
  Packages []*ResPackage
}

func ParseResTable(file string) *ResTable {
  if file == "" {
    return nil
  }
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return nil
  }
  rt := &ResTable{bytesReader: &bytesReader{Reader: bytes.NewReader(data), data: data}}
  rt.Header = rt.parseHeader()
  rt.PackageCount = rt.readUint32()
  rt.StrPool = rt.parseStrPool()
  if rt.PackageCount > 0 {
    rt.Packages = make([]*ResPackage, 0, rt.PackageCount)
    for i := uint32(0); i < rt.PackageCount; i++ {
      rt.Packages = append(rt.Packages, rt.parsePackage())
    }
  }
  return rt
}

func (rt *ResTable) parseHeader() *Header {
  return &Header{
    Type:       rt.readUint16(),
    HeaderSize: rt.readUint16(),
    Size:       rt.readUint32(),
  }
}

func (rt *ResTable) parseStrPool() *ResStrPool {
  s := rt.pos()
  header := rt.parseHeader()
  strCount := rt.readUint32()
  styleCount := rt.readUint32()
  flags := rt.readUint32()
  strStart := rt.readUint32()
  styleStart := rt.readUint32()
  strOffsets := rt.readUint32Array(strCount)
  styleOffsets := rt.readUint32Array(styleCount)

  var strs []string
  if strCount > 0 {
    e := s + header.Size
    if styleCount > 0 {
      e = s + styleStart
    }
    block := rt.slice(rt.pos(), e)
    strs = make([]string, strCount)
    if flags&0x0100 != 0 {
      // UTF-8
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str8(block, strOffsets[i])
      }
    } else {
      // UTF-16
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str16(block, strOffsets[i])
      }
    }
  }

  // todo 样式解析
  rt.Seek(int64(s+header.Size), io.SeekStart)

  return &ResStrPool{
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

func (rt *ResTable) parsePackage() *ResPackage {
  s := rt.pos()
  header := rt.parseHeader()
  id := rt.readUint32()
  // 包名是固定的256个字节，不足的会填充0，
  // UTF-16编码，每2个字节表示一个字符，所以字符之间会有0，需要去掉
  block := make([]byte, 0, 128)
  for _, v := range rt.readN(256) {
    if v != 0 {
      block = append(block, v)
    }
  }
  name := string(block)
  typeStrPoolStart := rt.readUint32()
  typeCount := rt.readUint32()
  keyStrPoolStart := rt.readUint32()
  keyCount := rt.readUint32()
  res0 := rt.readUint32()
  typeStrPool := rt.parseStrPool()
  keyStrPool := rt.parseStrPool()

  var typeSpecs []*ResTypeSpec
  var types []*ResType
  if typeCount > 0 {
    typeSpecs = make([]*ResTypeSpec, 0, typeCount)
    types = make([]*ResType, 0, typeCount)
    e := s + header.Size
    for rt.pos() < e {
      switch rt.readUint16() {
      case 0x0202:
        // Type Spec
        rt.unreadN(2)
        typeSpecs = append(typeSpecs, rt.parseTypeSpec())
      case 0x0201:
        // Type
        rt.unreadN(2)
        types = append(types, rt.parseType())
      }
    }
  }

  return &ResPackage{
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

func (rt *ResTable) parseTypeSpec() *ResTypeSpec {
  fmt.Println("parseTypeSpec1>>>>>", rt.pos())
  header := rt.parseHeader()
  id := rt.readUint8()
  res0 := rt.readUint8()
  res1 := rt.readUint16()
  entryCount := rt.readUint32()
  fmt.Println("parseTypeSpec2>>>>>", rt.pos())
  entryFlags := rt.readUint32Array(entryCount)
  fmt.Println("parseTypeSpec3>>>>>", rt.pos())
  return &ResTypeSpec{
    Header:     header,
    Id:         id,
    Res0:       res0,
    Res1:       res1,
    EntryCount: entryCount,
    EntryFlags: entryFlags,
  }
}

func (rt *ResTable) parseType() *ResType {
  p := rt.pos()
  header := rt.parseHeader()
  rt.Seek(int64(p+header.Size), io.SeekStart)
  return nil
  /*header := rt.parseHeader()
  id := rt.readUint8()
  res0 := rt.readUint8()
  res1 := rt.readUint16()
  entryCount := rt.readUint32()
  entryStart := rt.readUint32()
  config := rt.parseConfig()
  entryOffsets := rt.readUint32Array(entryCount)

  var entries []*ResEntry
  if entryCount > 0 {
    entries = make([]*ResEntry, 0, entryCount)
    for i := uint32(0); i < entryCount; i++ {
      if entryOffsets[i] != math.MaxUint32 {
        entries = append(entries, rt.parseEntry())
      }
    }
  }

  return &ResType{
    Header:       header,
    Id:           id,
    Res0:         res0,
    Res1:         res1,
    EntryCount:   entryCount,
    EntryStart:   entryStart,
    Config:       config,
    EntryOffsets: entryOffsets,
    Entries:      entries,
  }*/
}

func (rt *ResTable) parseConfig() *ResConfig {
  return &ResConfig{
    Size:                  rt.readUint32(),
    Mcc:                   rt.readUint16(),
    Mnc:                   rt.readUint16(),
    Language:              rt.readUint16(),
    Country:               rt.readUint16(),
    Orientation:           rt.readUint8(),
    Touchscreen:           rt.readUint8(),
    Density:               rt.readUint16(),
    Keyboard:              rt.readUint8(),
    Navigation:            rt.readUint8(),
    InputFlags:            rt.readUint8(),
    InputPad0:             rt.readUint8(),
    ScreenWidth:           rt.readUint16(),
    ScreenHeight:          rt.readUint16(),
    SdkVersion:            rt.readUint16(),
    MinorVersion:          rt.readUint16(),
    ScreenLayout:          rt.readUint8(),
    UiMode:                rt.readUint8(),
    SmallestScreenWidthDp: rt.readUint16(),
    ScreenWidthDp:         rt.readUint16(),
    ScreenHeightDp:        rt.readUint16(),
  }
}

func (rt *ResTable) parseEntry() *ResEntry {
  // todo
  return nil
}

func (rt *ResTable) parseValue() *ResValue {
  // todo
  return nil
}

type bytesReader struct {
  *bytes.Reader
  data []byte
}

func (r *bytesReader) pos() uint32 {
  return uint32(len(r.data) - r.Len())
}

func (r *bytesReader) slice(start, end uint32) []byte {
  r.Seek(int64(end), io.SeekStart)
  return r.data[start:end]
}

func (r *bytesReader) readN(n uint32) []byte {
  if n < 1 {
    return nil
  }
  ret := make([]byte, n)
  r.Read(ret)
  return ret
}

func (r *bytesReader) unreadN(n uint32) {
  if n < 1 {
    return
  }
  r.Seek(int64(-n), io.SeekCurrent)
  /*for i:=uint32(0) ;i < n; i++ {
    r.UnreadByte()
  }*/
}

func (r *bytesReader) readUint8() uint8 {
  b, _ := r.ReadByte()
  return uint8(b)
}

func (r *bytesReader) readUint16() uint16 {
  return conv.BytesToUint16L(r.readN(2))
}

func (r *bytesReader) readUint32() uint32 {
  return conv.BytesToUint32L(r.readN(4))
}

func (r *bytesReader) readUint32Array(count uint32) []uint32 {
  if count < 1 {
    return nil
  }
  ret := make([]uint32, count)
  for i := uint32(0); i < count; i++ {
    ret[i] = r.readUint32()
  }
  return ret
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
