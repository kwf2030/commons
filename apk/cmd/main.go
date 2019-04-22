package main

import (
  "flag"

  "github.com/kwf2030/commons/apk"
)

func main() {
  manifest := flag.String("m", "AndroidManifest.xml", "specify manifest file")
  debuggable := flag.Bool("d", false, "set debuggable=\"true\" for <application>")
  json := flag.Bool("j", false, "generate json file after decoding")
  flag.Parse()

  xml, e := apk.DecodeXmlFile(*manifest)
  if e != nil {
    panic(e)
  }

  if *debuggable {
    e = xml.AddAttr("android:debuggable", true, 3, 4, 0, func(tag *apk.Tag) bool {
      return tag.DecodedName == "application"
    })
    if e != nil {
      panic(e)
    }
    xml.AddResId(16842767, 4)
    e = xml.Marshal(*manifest)
    if e != nil {
      panic(e)
    }
  }

  if *json {
    xml, e = apk.DecodeXml(*manifest)
    if e != nil {
      panic(e)
    }
    e = xml.MarshalJSON(*manifest + ".json")
    if e != nil {
      panic(e)
    }
  }
}
