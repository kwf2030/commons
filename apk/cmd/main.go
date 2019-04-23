package main

import (
  "flag"

  "github.com/kwf2030/commons/apk"
)

func main() {
  m := flag.String("m", "AndroidManifest.xml", "")
  d := flag.Bool("d", false, "set debuggable=\"true\"")
  j := flag.Bool("j", false, "dump json")
  flag.Parse()

  xml, e := apk.DecodeXmlFile(*m)
  if e != nil {
    panic(e)
  }

  if *d {
    e = xml.AddAttr("android:debuggable", true, 3, 4, 0, func(tag *apk.Tag) bool {
      return tag.DecodedName == "application"
    })
    if e != nil {
      panic(e)
    }
    xml.AddResId(16842767, 4)
    e = xml.MarshalFile(*m)
    if e != nil {
      panic(e)
    }
  }

  if *j {
    xml, e = apk.DecodeXmlFile(*m)
    if e != nil {
      panic(e)
    }
    e = xml.Dump(*m + ".json")
    if e != nil {
      panic(e)
    }
  }
}
