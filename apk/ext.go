package main

import "sort"

type ResTableEntry2 struct {
  Id        int
  PkgName   string
  TypeName  string
  KeyName   string
  Value     string
  ParentRef int
  Values    map[int]string
}

func (rt *ResTable) CollectResTableEntries() []*ResTableEntry2 {
  ret := make([]*ResTableEntry2, 0, 102400)
  for _, pkg := range rt.Packages {
    for _, tp := range pkg.Types {
      for i, entry := range tp.Entries {
        item := &ResTableEntry2{
          Id:       int(pkg.Id)<<24 | int(tp.Id)<<16 | i,
          PkgName:  pkg.Name,
          TypeName: pkg.TypeStrPool.Strs[tp.Id-1],
          KeyName:  pkg.KeyStrPool.Strs[entry.Key],
        }
        if entry.Flags&0x0001 == 0 {
          // item.Value = entry.Value
        } else {
          item.ParentRef = int(entry.ParentRef)
          item.Values = make(map[int]string, entry.Count)
          for k, v := range entry.Values {
            // item.Values[int(key)] = val
          }
        }
        ret = append(ret, item)
      }
    }
  }
  sort.Slice(ret, func(i, j int) bool {
    return ret[i].Id < ret[j].Id
  })
  return ret
}

func (rt *ResTable) parseData(dataType int, data int, pool *ResTableStrPool) string {
  switch dataType {
  case 3:
    // 字符串
  case 16:
    // 数字
  case 18:
    // 布尔
  }
  return ""
}

func GetResTableEntriesMap(entries []*ResTableEntry2) map[int]string {
  l := len(entries)
  if l == 0 {
    return nil
  }
  ret := make(map[int]string, l)
  for _, entry := range entries {
    ret[entry.Id] = entry.TypeName + "/" + entry.KeyName
  }
  return ret
}
