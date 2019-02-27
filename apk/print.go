package main

import (
  "errors"
  "fmt"
  "runtime"
  "strconv"
)

func main() {
  showXml()
  // showResTable()
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
  fmt.Println("MagicNumber:", xml.MagicNumber)
  fmt.Println("FileSize:", xml.FileSize)
  fmt.Println("========String Pool========")
  fmt.Println("Type:", xml.StrPool.Type)
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
  fmt.Println("Size:", xml.ResId.Size)
  fmt.Println("Count", len(xml.ResId.Ids))
  fmt.Println("========Namespace========")
  fmt.Println("Type:", xml.Namespace.Type)
  fmt.Println("Size:", xml.Namespace.Size)
  fmt.Println("LineNumber", xml.Namespace.LineNumber)
  fmt.Println("Prefix", xml.Namespace.Prefix, xml.StrPool.Strs[xml.Namespace.Prefix])
  fmt.Println("Uri", xml.Namespace.Uri, xml.StrPool.Strs[xml.Namespace.Uri])
  fmt.Println("========Tag========")
  fmt.Println("Type:", xml.Tag.Type)
  fmt.Println("Size:", xml.Tag.Size)
  fmt.Println("LineNumber", xml.Tag.LineNumber)
  fmt.Println("NamespaceUri", xml.Tag.NamespaceUri)
  fmt.Println("Name", xml.Tag.Name)
  fmt.Println("Flags", xml.Tag.Flags)
  fmt.Println("AttrCount", xml.Tag.AttrCount)
  fmt.Println("ClassAttr", xml.Tag.ClassAttr)
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
    fmt.Println("EntryConfigSize:", v.EntryConfig.Size)
    for j := 0; j < 5; j++ {
      if len(v.Entries) <= j {
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
