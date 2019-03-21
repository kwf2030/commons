package apk

import (
  "math"
  "strconv"
)

type ResTableEntry2 struct {
  Ori      *ResTableEntry
  Id       uint32
  PkgName  string
  TypeName string
  KeyName  string
  Name     string
  Value    string
  Values   map[uint32]string
}

type ResTable2 struct {
  Ori     *ResTable
  Entries map[uint32]*ResTableEntry2
}

func NewResTable2(rt *ResTable) *ResTable2 {
  if rt == nil {
    return nil
  }
  ret := &ResTable2{Ori: rt}
  ret.Entries = ret.collectEntries()
  return ret
}

func (rt2 *ResTable2) collectEntries() map[uint32]*ResTableEntry2 {
  ret := make(map[uint32]*ResTableEntry2, 40960)
  for _, pkg := range rt2.Ori.Packages {
    for _, tp := range pkg.Types {
      for i, entry := range tp.Entries {
        if entry == nil {
          continue
        }
        item := &ResTableEntry2{
          Ori:      entry,
          Id:       pkg.Id<<24 | uint32(tp.Id)<<16 | uint32(i),
          PkgName:  pkg.Name,
          TypeName: pkg.TypeStrPool.Strs[tp.Id-1],
          KeyName:  pkg.KeyStrPool.Strs[entry.Key],
        }
        item.Name = item.TypeName + "/" + item.KeyName
        if entry.Flags&0x0001 == 0 {
          item.Value = rt2.parseData(entry.Value.DataType, entry.Value.Data)
        } else {
          if entry.Count > 0 && entry.Count < math.MaxUint32 {
            item.Values = make(map[uint32]string, entry.Count)
            for k, v := range entry.Values {
              item.Values[k] = rt2.parseData(v.DataType, v.Data)
            }
          }
        }
        ret[item.Id] = item
      }
    }
  }
  return ret
}

func (rt2 *ResTable2) parseData(dataType uint8, data uint32) string {
  switch dataType {
  case 3:
    if data < math.MaxUint32 {
      return rt2.Ori.StrPool.Strs[data]
    }
  case 16:
    return strconv.Itoa(int(data))
  case 18:
    if data == 0 {
      return "false"
    }
    return "true"
  }
  return ""
}
