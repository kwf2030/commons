package main

import (
  "math"
  "strconv"
)

type ResTableEntry2 struct {
  Id        int
  PkgName   string
  TypeName  string
  KeyName   string
  Name      string
  Value     string
  ParentRef int
  Values    map[int]string
}

type ResTable2 struct {
  *ResTable
  Entries map[int]*ResTableEntry2
}

func NewResTable2(rt *ResTable) *ResTable2 {
  ret := &ResTable2{ResTable: rt}
  ret.Entries = ret.CollectEntries()
  return ret
}

func (rt2 *ResTable2) CollectEntries() map[int]*ResTableEntry2 {
  ret := make(map[int]*ResTableEntry2, 40960)
  for _, pkg := range rt2.Packages {
    for _, tp := range pkg.Types {
      for i, entry := range tp.Entries {
        if entry == nil {
          continue
        }
        item := &ResTableEntry2{
          Id:       int(pkg.Id)<<24 | int(tp.Id)<<16 | i,
          PkgName:  pkg.Name,
          TypeName: pkg.TypeStrPool.Strs[tp.Id-1],
          KeyName:  pkg.KeyStrPool.Strs[entry.Key],
        }
        item.Name = item.TypeName + "/" + item.KeyName
        if entry.Flags&0x0001 == 0 {
          item.Value = rt2.parseData(entry.Value)
        } else {
          item.ParentRef = int(entry.ParentRef)
          if entry.Count > 0 && entry.Count < math.MaxUint32 {
            item.Values = make(map[int]string, entry.Count)
            for k, v := range entry.Values {
              item.Values[int(k)] = rt2.parseData(v)
            }
          }
        }
        ret[item.Id] = item
      }
    }
  }
  return ret
}

func (rt2 *ResTable2) parseData(value *ResTableValue) string {
  switch value.DataType {
  case 3:
    if value.Data < math.MaxUint32 {
      return rt2.StrPool.Strs[value.Data]
    }
  case 16:
    return strconv.Itoa(int(value.Data))
  case 18:
    if value.Data == 0 {
      return "false"
    }
    return "true"
  }
  return ""
}
