package apk

import (
  "bytes"
  "io/ioutil"
  "math"
)

type ResTableHeader struct {
  Type       uint16
  HeaderSize uint16
  Size       uint32
}

type ResTableStrPool struct {
  // Chunk的起始和结束位置，非协议字段
  ChunkStart, ChunkEnd uint32

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

  // 字符串偏移数组（相对Strs），长度为StrCount
  StrOffsets []uint32

  // 字符串样式偏移数组（相对Styles），长度为StyleCount
  StyleOffsets []uint32

  // 字符串，长度为StrCount，每个字符串前两个字节为该字符串长度，
  // 若是UTF-8编码，以0x00（1个字节）作为结束符，
  // 若是UTF-16编码，以0x0000（2个字节）作为结束符
  Strs []string

  Styles []byte
}

type ResTablePackage struct {
  // Chunk的起始和结束位置，非协议字段
  ChunkStart, ChunkEnd uint32

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

  // 资源类型字符串池（UTF-8）
  TypeStrPool *ResTableStrPool

  // 资源项名称字符串池（UTF-8）
  KeyStrPool *ResTableStrPool

  TypeSpecs []*ResTableTypeSpec

  Types []*ResTableType
}

type ResTableTypeSpec struct {
  // Chunk的起始和结束位置，非协议字段
  ChunkStart, ChunkEnd uint32

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
  // Chunk的起始和结束位置，非协议字段
  ChunkStart, ChunkEnd uint32

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
  Config *ResTableConfig

  // 资源项偏移数组，长度为EntryCount
  EntryOffsets []uint32

  // 资源项，长度为EntryCount
  Entries []*ResTableEntry
}

type ResTableConfig struct {
  // Chunk的起始和结束位置，非协议字段
  ChunkStart, ChunkEnd uint32

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

  // 剩余未解析的字节
  Res0 []byte
}

type ResTableEntry struct {
  // Chunk的起始和结束位置，非协议字段
  ChunkStart, ChunkEnd uint32

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
  // Chunk的起始和结束位置，非协议字段
  ChunkStart, ChunkEnd uint32

  Size     uint16
  Res0     uint8
  DataType uint8
  Data     uint32
}

type ResTable struct {
  // 非协议字段
  *bytesReader `json:"-"`

  // Chunk的起始和结束位置，非协议字段
  ChunkStart, ChunkEnd uint32

  *ResTableHeader

  // 资源包个数，通常一个app只有一个资源包
  PackageCount uint32

  // 全局字符串池（UTF-8）
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
  rt := &ResTable{bytesReader: &bytesReader{Reader: bytes.NewReader(data), data: data}, ChunkStart: 0}
  rt.ResTableHeader = rt.parseHeader()
  rt.PackageCount = rt.readUint32()
  rt.StrPool = rt.parseStrPool()
  if rt.PackageCount > 0 && rt.PackageCount < math.MaxUint32 {
    rt.Packages = make([]*ResTablePackage, rt.PackageCount)
    for i := uint32(0); i < rt.PackageCount; i++ {
      rt.Packages[i] = rt.parsePackage()
    }
  }
  rt.ChunkEnd = rt.Size
  return rt
}

func (rt *ResTable) parseHeader() *ResTableHeader {
  return &ResTableHeader{
    Type:       rt.readUint16(),
    HeaderSize: rt.readUint16(),
    Size:       rt.readUint32(),
  }
}

func (rt *ResTable) parseStrPool() *ResTableStrPool {
  chunkStart := rt.pos()
  header := rt.parseHeader()
  strCount := rt.readUint32()
  styleCount := rt.readUint32()
  flags := rt.readUint32()
  strStart := rt.readUint32()
  styleStart := rt.readUint32()
  strOffsets := rt.readUint32Array(strCount)
  styleOffsets := rt.readUint32Array(styleCount)

  var strs []string
  if strCount > 0 && styleCount < math.MaxUint32 {
    end := chunkStart + header.Size
    if styleCount > 0 && styleCount < math.MaxUint32 {
      end = chunkStart + styleStart
    }
    pool := rt.slice(rt.pos(), end)
    strs = make([]string, strCount)
    if flags&0x0100 != 0 {
      for i := uint32(0); i < strCount; i++ {
        strs[i] = string(str8(pool, strOffsets[i]))
      }
    } else {
      for i := uint32(0); i < strCount; i++ {
        strs[i] = string(str16(pool, strOffsets[i]))
      }
    }
  }

  // todo 样式解析
  var styles []byte
  if styleCount > 0 && styleCount < math.MaxUint32 {
    styles = rt.slice(chunkStart+styleStart, chunkStart+header.Size)
  }

  return &ResTableStrPool{
    ChunkStart:     chunkStart,
    ChunkEnd:       chunkStart + header.Size,
    ResTableHeader: header,
    StrCount:       strCount,
    StyleCount:     styleCount,
    Flags:          flags,
    StrStart:       strStart,
    StyleStart:     styleStart,
    StrOffsets:     strOffsets,
    StyleOffsets:   styleOffsets,
    Strs:           strs,
    Styles:         styles,
  }
}

func (rt *ResTable) parsePackage() *ResTablePackage {
  chunkStart := rt.pos()
  header := rt.parseHeader()
  id := rt.readUint32()
  // 包名是固定的256个字节（UTF-16编码），不足的会填充0，
  // 需要去掉多余的0
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
  typeStrPool := rt.parseStrPool()
  keyStrPool := rt.parseStrPool()

  var typeSpecs []*ResTableTypeSpec
  var types []*ResTableType
  if typeCount > 0 && typeCount < math.MaxUint32 {
    typeSpecs = make([]*ResTableTypeSpec, 0, typeCount)
    types = make([]*ResTableType, 0, 256)
  }
  chunkEnd := chunkStart + header.Size
  for rt.pos() < chunkEnd {
    switch rt.readUint16() {
    case 514:
      rt.unreadN(2)
      typeSpecs = append(typeSpecs, rt.parseTypeSpec())
    case 513:
      rt.unreadN(2)
      types = append(types, rt.parseType())
    }
  }

  return &ResTablePackage{
    ChunkStart:       chunkStart,
    ChunkEnd:         chunkEnd,
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

func (rt *ResTable) parseTypeSpec() *ResTableTypeSpec {
  chunkStart := rt.pos()
  header := rt.parseHeader()
  id := rt.readUint8()
  res0 := rt.readUint8()
  res1 := rt.readUint16()
  entryCount := rt.readUint32()
  entryFlags := rt.readUint32Array(entryCount)
  return &ResTableTypeSpec{
    ChunkStart:     chunkStart,
    ChunkEnd:       chunkStart + header.Size,
    ResTableHeader: header,
    Id:             id,
    Res0:           res0,
    Res1:           res1,
    EntryCount:     entryCount,
    EntryFlags:     entryFlags,
  }
}

func (rt *ResTable) parseType() *ResTableType {
  chunkStart := rt.pos()
  header := rt.parseHeader()
  id := rt.readUint8()
  res0 := rt.readUint8()
  res1 := rt.readUint16()
  entryCount := rt.readUint32()
  entryStart := rt.readUint32()
  config := rt.parseConfig()
  entryOffsets := rt.readUint32Array(entryCount)

  var entries []*ResTableEntry
  if entryCount > 0 && entryCount < math.MaxUint32 {
    entries = make([]*ResTableEntry, entryCount)
    for i := uint32(0); i < entryCount; i++ {
      // 遍历时注意entries的元素可能为nil
      if entryOffsets[i] > 0 && entryOffsets[i] < math.MaxUint32 {
        entries[i] = rt.parseEntry()
      }
    }
  }

  return &ResTableType{
    ChunkStart:     chunkStart,
    ChunkEnd:       chunkStart + header.Size,
    ResTableHeader: header,
    Id:             id,
    Res0:           res0,
    Res1:           res1,
    EntryCount:     entryCount,
    EntryStart:     entryStart,
    Config:         config,
    EntryOffsets:   entryOffsets,
    Entries:        entries,
  }
}

func (rt *ResTable) parseConfig() *ResTableConfig {
  // ResTableConfig一共56个字节，
  // 目前只解析了36个字节，未解析的Res0是20个字节
  chunkStart := rt.pos()
  size := rt.readUint32()
  ret := &ResTableConfig{
    ChunkStart:            chunkStart,
    ChunkEnd:              chunkStart + size,
    Size:                  size,
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
    Res0:                  rt.slice(rt.pos(), chunkStart+size),
  }
  return ret
}

func (rt *ResTable) parseEntry() *ResTableEntry {
  chunkStart := rt.pos()
  size := rt.readUint16()
  flags := rt.readUint16()
  key := rt.readUint32()

  if flags&0x0001 == 0 {
    return &ResTableEntry{
      ChunkStart: chunkStart,
      Size:       size,
      Flags:      flags,
      Key:        key,
      Value:      rt.parseValue(),
      ChunkEnd:   rt.pos(),
    }
  }

  parentRef := rt.readUint32()
  count := rt.readUint32()
  var values map[uint32]*ResTableValue
  if count > 0 && count < math.MaxUint32 {
    values = make(map[uint32]*ResTableValue, count)
    for i := uint32(0); i < count; i++ {
      values[rt.readUint32()] = rt.parseValue()
    }
  }
  return &ResTableEntry{
    ChunkStart: chunkStart,
    Size:       size,
    Flags:      flags,
    Key:        key,
    ParentRef:  parentRef,
    Count:      count,
    Values:     values,
    ChunkEnd:   rt.pos(),
  }
}

func (rt *ResTable) parseValue() *ResTableValue {
  return &ResTableValue{
    ChunkStart: rt.pos(),
    Size:       rt.readUint16(),
    Res0:       rt.readUint8(),
    DataType:   rt.readUint8(),
    Data:       rt.readUint32(),
    ChunkEnd:   rt.pos(),
  }
}
