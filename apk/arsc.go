package main

import (
  "io/ioutil"
  "math"

  "github.com/kwf2030/commons/conv"
)

// Header Size: 28
// Size: 28+StrCount*4+StyleCount*4+Strs+Styles


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



// Header Size: 288
// Size: 288+TypeStrPool+KeyStrPool+Types+TypeSpecs


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

  var entries []ResEntry
  if entryCount > 0 {
    entries = make([]ResEntry, 0, entryCount)
    offset += entryStart
    for i := uint32(0); i < entryCount; i++ {
      if entryOffsets[i] != math.MaxUint32 {
        e := parseEntry(data, offset)
        entries = append(entries, e)
        offset += uint32(e.Size)
      }
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
    Entries:      entries,
  }
}

// Size: 56


func parseConfig(data []byte, offset uint32) ResConfig {
  return
}

