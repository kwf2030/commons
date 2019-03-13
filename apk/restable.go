package main

import (
  "bytes"
  "io"
  "io/ioutil"
  "math"
)

type ResTableHeader struct {
  Type       uint16
  HeaderSize uint16
  Size       uint32
}

type ResTableStrPool struct {
  *ResTableHeader

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

  // 字符串，长度为StrCount，每个字符串前两个字节为该字符串长度，
  // 若是UTF-8编码，以0x00（1个字节）作为结束符，
  // 若是UTF-16编码，以0x0000（2个字节）作为结束符
  Strs []string

  Styles []string
}

type ResTablePackage struct {
  *ResTableHeader

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
  TypeStrPool *ResTableStrPool

  // 资源项名称字符串池
  KeyStrPool *ResTableStrPool

  TypeSpecs []*ResTableTypeSpec

  Types []*ResTableType
}

type ResTableTypeSpec struct {
  *ResTableHeader

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

type ResTableType struct {
  *ResTableHeader

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
  EntryConfig *ResTableEntryConfig

  // 资源项偏移数组，长度为EntryCount
  EntryOffsets []uint32

  // 资源项，长度为EntryCount
  Entries []*ResTableEntry
}

type ResTableEntryConfig struct {
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

  // 剩余20个字节未解析
  Res0 []byte
}

type ResTableEntry struct {
  Size uint16
  // Flags&0x0001==0，Value有值，
  // 否则，ParentRef/Count/Values有值
  Flags     uint16
  Key       uint32
  Value     *ResTableValue
  ParentRef uint32
  Count     uint32
  Values    map[uint32]*ResTableValue
}

type ResTableValue struct {
  Size     uint16
  Res0     uint8
  DataType uint8
  Data     uint32
}

type ResTable struct {
  *bytesReader

  *ResTableHeader

  // 资源包个数，通常一个app只有一个资源包
  PackageCount uint32

  // 全局字符串池
  StrPool *ResTableStrPool

  // 资源包，长度为PackageCount
  Packages []*ResTablePackage
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
  rt.ResTableHeader = rt.parseResTableHeader()
  rt.PackageCount = rt.readUint32()
  rt.StrPool = rt.parseResTableStrPool()
  if rt.PackageCount > 0 && rt.PackageCount < math.MaxUint32 {
    rt.Packages = make([]*ResTablePackage, rt.PackageCount)
    for i := uint32(0); i < rt.PackageCount; i++ {
      rt.Packages[i] = rt.parseResTablePackage()
    }
  }
  return rt
}

func (rt *ResTable) parseResTableHeader() *ResTableHeader {
  return &ResTableHeader{
    Type:       rt.readUint16(),
    HeaderSize: rt.readUint16(),
    Size:       rt.readUint32(),
  }
}

func (rt *ResTable) parseResTableStrPool() *ResTableStrPool {
  s := rt.pos()
  header := rt.parseResTableHeader()
  strCount := rt.readUint32()
  styleCount := rt.readUint32()
  flags := rt.readUint32()
  strStart := rt.readUint32()
  styleStart := rt.readUint32()
  strOffsets := rt.readUint32Array(strCount)
  styleOffsets := rt.readUint32Array(styleCount)

  var strs []string
  if strCount > 0 && styleCount < math.MaxUint32 {
    e := s + header.Size
    if styleCount > 0 && styleCount < math.MaxUint32 {
      e = s + styleStart
    }
    block := rt.slice(rt.pos(), e)
    strs = make([]string, strCount)
    if flags&0x0100 != 0 {
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str8(block, strOffsets[i])
      }
    } else {
      for i := uint32(0); i < strCount; i++ {
        strs[i] = str16(block, strOffsets[i])
      }
    }
  }

  // todo 样式解析
  rt.Seek(int64(s+header.Size), io.SeekStart)

  return &ResTableStrPool{
    ResTableHeader: header,
    StrCount:       strCount,
    StyleCount:     styleCount,
    Flags:          flags,
    StrStart:       strStart,
    StyleStart:     styleStart,
    StrOffsets:     strOffsets,
    StyleOffsets:   styleOffsets,
    Strs:           strs,
    Styles:         nil,
  }
}

func (rt *ResTable) parseResTablePackage() *ResTablePackage {
  s := rt.pos()
  header := rt.parseResTableHeader()
  id := rt.readUint32()
  // 包名是固定的256个字节，不足的会填充0，
  // UTF-16编码，每2个字节表示一个字符，所以字符之间会有0，需要去掉
  arr := make([]byte, 0, 128)
  for _, v := range rt.readN(256) {
    if v != 0 {
      arr = append(arr, v)
    }
  }
  name := string(arr)
  typeStrPoolStart := rt.readUint32()
  typeCount := rt.readUint32()
  keyStrPoolStart := rt.readUint32()
  keyCount := rt.readUint32()
  res0 := rt.readUint32()
  typeStrPool := rt.parseResTableStrPool()
  keyStrPool := rt.parseResTableStrPool()

  var typeSpecs []*ResTableTypeSpec
  if typeCount > 0 && typeCount < math.MaxUint32 {
    typeSpecs = make([]*ResTableTypeSpec, 0, typeCount)
  }
  var types []*ResTableType
  if keyCount > 0 && keyCount < math.MaxUint32 {
    types = make([]*ResTableType, 0, keyCount)
  }
  e := s + header.Size
  for rt.pos() < e {
    switch rt.readUint16() {
    case 514:
      rt.unreadN(2)
      typeSpecs = append(typeSpecs, rt.parseResTableTypeSpec())
    case 513:
      rt.unreadN(2)
      types = append(types, rt.parseResTableType())
    }
  }

  return &ResTablePackage{
    ResTableHeader:   header,
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

func (rt *ResTable) parseResTableTypeSpec() *ResTableTypeSpec {
  header := rt.parseResTableHeader()
  id := rt.readUint8()
  res0 := rt.readUint8()
  res1 := rt.readUint16()
  entryCount := rt.readUint32()
  entryFlags := rt.readUint32Array(entryCount)
  return &ResTableTypeSpec{
    ResTableHeader: header,
    Id:             id,
    Res0:           res0,
    Res1:           res1,
    EntryCount:     entryCount,
    EntryFlags:     entryFlags,
  }
}

func (rt *ResTable) parseResTableType() *ResTableType {
  header := rt.parseResTableHeader()
  id := rt.readUint8()
  res0 := rt.readUint8()
  res1 := rt.readUint16()
  entryCount := rt.readUint32()
  entryStart := rt.readUint32()
  entryConfig := rt.parseResTableEntryConfig()
  entryOffsets := rt.readUint32Array(entryCount)

  var entries []*ResTableEntry
  if entryCount > 0 && entryCount < math.MaxUint32 {
    entries = make([]*ResTableEntry, entryCount)
    for i := uint32(0); i < entryCount; i++ {
      if entryOffsets[i] > 0 && entryOffsets[i] < math.MaxUint32 {
        entries[i] = rt.parseResTableEntry()
      }
    }
  }

  return &ResTableType{
    ResTableHeader: header,
    Id:             id,
    Res0:           res0,
    Res1:           res1,
    EntryCount:     entryCount,
    EntryStart:     entryStart,
    EntryConfig:    entryConfig,
    EntryOffsets:   entryOffsets,
    Entries:        entries,
  }
}

func (rt *ResTable) parseResTableEntryConfig() *ResTableEntryConfig {
  // 76个字节
  ret := &ResTableEntryConfig{
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
    Res0:                  rt.readN(20),
  }
  return ret
}

func (rt *ResTable) parseResTableEntry() *ResTableEntry {
  size := rt.readUint16()
  flags := rt.readUint16()
  key := rt.readUint32()

  if flags&0x0001 == 0 {
    return &ResTableEntry{
      Size:  size,
      Flags: flags,
      Key:   key,
      Value: rt.parseResTableValue(),
    }
  }

  parentRef := rt.readUint32()
  count := rt.readUint32()
  var values map[uint32]*ResTableValue
  if count > 0 && count < math.MaxUint32 {
    values = make(map[uint32]*ResTableValue, count)
    for i := uint32(0); i < count; i++ {
      values[rt.readUint32()] = rt.parseResTableValue()
    }
  }
  return &ResTableEntry{
    Size:      size,
    Flags:     flags,
    Key:       key,
    ParentRef: parentRef,
    Count:     count,
    Values:    values,
  }
}

func (rt *ResTable) parseResTableValue() *ResTableValue {
  return &ResTableValue{
    Size:     rt.readUint16(),
    Res0:     rt.readUint8(),
    DataType: rt.readUint8(),
    Data:     rt.readUint32(),
  }
}
