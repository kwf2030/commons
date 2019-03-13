package main

import (
  "errors"
  "fmt"
  "math"
  "runtime"
  "strconv"
)

func main() {
  // showXml()
  showResTable()
  //showResTable2()
}

func showResTable2() {
  var file string
  switch runtime.GOOS {
  case "windows":
    file = "C:\\Users\\WangFeng\\Desktop\\resources.arsc"
  case "linux":
    file = "/home/wangfeng/workspace/wechat/raw/resources.arsc"
  default:
    panic(errors.New("os not supported"))
  }
  rt := ParseResTable(file)
  rt2 := NewResTable2(rt)
  for k, v := range rt2.Entries {
    fmt.Println(k, v)
  }
}

func showXml() {
  var file string
  switch runtime.GOOS {
  case "windows":
    file = "C:\\Users\\WangFeng\\Desktop\\AndroidManifest.xml"
  case "linux":
    file = "/home/wangfeng/workspace/wechat/raw/AndroidManifest.xml"
  default:
    panic(errors.New("os not supported"))
  }
  xml := ParseXml(file)
  printXmlInfo(xml)
}

func printXmlInfo(xml *Xml) {
  fmt.Println("====Xml====")
  fmt.Println("Version:", xml.Type)
  fmt.Println("HeaderSize:", xml.HeaderSize)
  fmt.Println("Size:", xml.Size)
  fmt.Println("NamespaceCount:", len(xml.Namespaces))
  fmt.Println("TagCount:", len(xml.Tags))
  fmt.Println("========String Pool========")
  fmt.Println("Type:", xml.StrPool.Type)
  fmt.Println("HeaderSize:", xml.StrPool.HeaderSize)
  fmt.Println("Size:", xml.StrPool.Size)
  fmt.Println("StrCount:", xml.StrPool.StrCount)
  fmt.Println("StyleCount:", xml.StrPool.StyleCount)
  fmt.Println("Flags:", xml.StrPool.Flags)
  fmt.Println("StrStart:", xml.StrPool.StrStart)
  fmt.Println("StyleStart:", xml.StrPool.StyleStart)
  /*for i := 0; i < 10; i++ {
    fmt.Println("Sting Pool", i, xml.StrPool.Strs[i])
  }*/
  fmt.Println("========Resource Id========")
  fmt.Println("Type:", xml.ResId.Type)
  fmt.Println("HeaderSize:", xml.ResId.HeaderSize)
  fmt.Println("Size:", xml.ResId.Size)
  fmt.Println("Count:", len(xml.ResId.Ids))
  for i := 0; i < 10; i++ {
    if len(xml.Namespaces) <= i {
      continue
    }
    fmt.Println("========Namespace[" + strconv.Itoa(i) + "]========")
    fmt.Println("Type:", xml.Namespaces[i].Type)
    fmt.Println("HeaderSize:", xml.Namespaces[i].HeaderSize)
    fmt.Println("Size:", xml.Namespaces[i].Size)
    fmt.Println("LineNumber:", xml.Namespaces[i].LineNumber)
    fmt.Println("Prefix:", xml.Namespaces[i].Prefix, xml.StrPool.Strs[xml.Namespaces[i].Prefix])
    fmt.Println("Uri:", xml.Namespaces[i].Uri, xml.StrPool.Strs[xml.Namespaces[i].Uri])
  }
  for i := 0; i < 10; i++ {
    if len(xml.Tags) <= i {
      continue
    }
    fmt.Println("========Tag[" + strconv.Itoa(i) + "]========")
    fmt.Println("Type:", xml.Tags[i].Type)
    fmt.Println("HeaderSize:", xml.Tags[i].HeaderSize)
    fmt.Println("Size:", xml.Tags[i].Size)
    fmt.Println("LineNumber:", xml.Tags[i].LineNumber)
    fmt.Println("NamespaceUri:", xml.Tags[i].NamespaceUri)
    name := xml.Tags[i].Name
    if name < math.MaxUint32 {
      fmt.Println("Name:", name, xml.StrPool.Strs[name])
    } else {
      fmt.Println("Name:", name)
    }
    fmt.Println("AttrStart:", xml.Tags[i].AttrStart)
    fmt.Println("AttrSize:", xml.Tags[i].AttrSize)
    fmt.Println("AttrCount:", xml.Tags[i].AttrCount)
    fmt.Println("IdIndex:", xml.Tags[i].IdIndex)
    fmt.Println("ClassIndex:", xml.Tags[i].ClassIndex)
    fmt.Println("StyleIndex:", xml.Tags[i].StyleIndex)
    for i, v := range xml.Tags[i].Attrs {
      fmt.Println("============Attrs[" + strconv.Itoa(i) + "]============")
      ns := v.Namespace
      if ns < math.MaxUint32 {
        fmt.Println("Namespace:", ns, xml.StrPool.Strs[ns])
      } else {
        fmt.Println("Namespace:", ns)
      }
      name2 := v.Name
      if name2 < math.MaxUint32 {
        fmt.Println("Name:", name2, xml.StrPool.Strs[name2])
      } else {
        fmt.Println("Name:", name2)
      }
      fmt.Println("RawValue:", v.RawValue)
      fmt.Println("ValueSize:", v.ValueSize)
      dt := v.DataType
      fmt.Println("DataType:", dt)
      switch data := v.Data; dt {
      case 3:
        if data < math.MaxUint32 {
          fmt.Println("Data:", data, xml.StrPool.Strs[data])
        } else {
          fmt.Println("Data:", data)
        }
      case 16:
        fmt.Println("Data:", data)
      case 18:
        fmt.Println("Data:", data != 0)
      }
    }
  }
}

func showResTable() {
  var file string
  switch runtime.GOOS {
  case "windows":
    file = "C:\\Users\\WangFeng\\Desktop\\resources.arsc"
  case "linux":
    file = "/home/wangfeng/workspace/wechat/raw/resources.arsc"
  default:
    panic(errors.New("os not supported"))
  }
  rt := ParseResTable(file)
  printResTableInfo(rt)
}

func printResTableInfo(rt *ResTable) {
  fmt.Println("====Table====")
  fmt.Println("Type:", rt.Type)
  fmt.Println("Size:", rt.Size)
  fmt.Println("HeaderSize:", rt.HeaderSize)
  fmt.Println("========String Pool========")
  fmt.Println("Type:", rt.StrPool.Type)
  fmt.Println("Size:", rt.StrPool.Size)
  fmt.Println("HeaderSize:", rt.StrPool.HeaderSize)
  fmt.Println("Flags:", rt.StrPool.Flags)
  fmt.Println("StrCount:", rt.StrPool.StrCount)
  fmt.Println("StrStart:", rt.StrPool.StrStart)
  fmt.Println("StyleCount:", rt.StrPool.StyleCount)
  fmt.Println("StyleStart:", rt.StrPool.StyleStart)
  /*for i := 0; i < 10; i++ {
    fmt.Println("Sting Pool", i, table.StrPool.Strs[i])
  }*/
  fmt.Println("========Package========")
  fmt.Println("Type:", rt.Packages[0].Type)
  fmt.Println("Size:", rt.Packages[0].Size)
  fmt.Println("HeaderSize:", rt.Packages[0].HeaderSize)
  fmt.Println("Types:", rt.Packages[0].TypeStrPool.Strs)
  fmt.Printf("TypeCount/TypeStrCount: %d/%d\n", rt.Packages[0].TypeCount, rt.Packages[0].TypeStrPool.StrCount)
  fmt.Println("TypeStrPoolStart:", rt.Packages[0].TypeStrPoolStart)
  fmt.Printf("KeyCount/KeyStrCount: %d/%d\n", rt.Packages[0].KeyCount, rt.Packages[0].KeyStrPool.StrCount)
  fmt.Println("KeyStrPoolStart:", rt.Packages[0].KeyStrPoolStart)
  /*for i := 0; i < 10; i++ {
    fmt.Println("Type Sting Pool", i, table.Packages[0].TypeStrPool.Strs[i])
  }*/
  /*for i := 0; i < 10; i++ {
    fmt.Println("Key Sting Pool", i, table.Packages[0].KeyStrPool.Strs[i])
  }*/
  for i, v := range rt.Packages[0].TypeSpecs {
    if v == nil {
      continue
    }
    fmt.Println("============Type Spec[" + strconv.Itoa(i) + "]============")
    fmt.Println("Type:", v.Type)
    fmt.Println("Size:", v.Size)
    fmt.Println("HeaderSize:", v.HeaderSize)
    fmt.Println("Id:", v.Id)
    fmt.Println("EntryCount:", v.EntryCount)
  }
  for i, v := range rt.Packages[0].Types {
    if v == nil {
      continue
    }
    fmt.Println("============Types[" + strconv.Itoa(i) + "]============")
    fmt.Println("Type:", v.Type)
    fmt.Println("Size:", v.Size)
    fmt.Println("HeaderSize:", v.HeaderSize)
    fmt.Println("Id:", v.Id)
    fmt.Println("EntryCount:", v.EntryCount)
    fmt.Println("EntryStart:", v.EntryStart)
    fmt.Println("ConfigSize:", v.Config.Size)
    for j := 0; j < 5; j++ {
      if len(v.Entries) <= j || v.Entries[j] == nil {
        continue
      }
      fmt.Println("================Entries[" + strconv.Itoa(j) + "]==============")
      fmt.Println("Size:", v.Entries[j].Size)
      fmt.Println("Flags:", v.Entries[j].Flags)
      fmt.Println("Key:", v.Entries[j].Key)
      if v.Entries[j].Flags&0x0001 == 0 {
        fmt.Println("Value:", v.Entries[j].Value)
      } else {
        fmt.Println("ParentRef:", v.Entries[j].ParentRef)
        fmt.Println("Count:", v.Entries[j].Count)
      }
    }
  }
}
