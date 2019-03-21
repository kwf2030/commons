package main

import (
  "flag"

  "github.com/kwf2030/commons/apk"
)

func main() {
  manifest := flag.String("m", "AndroidManifest.xml", "specify manifest file")
  debuggable := flag.Bool("d", false, "set debuggable=\"true\" for application node")
  json := flag.Bool("j", false, "generate json file for manifest")
  flag.Parse()

  xml, e := apk.DecodeXml(*manifest)
  if e != nil {
    panic(e)
  }

  if *debuggable {
    xml.AddAttr("android:debuggable", true, func(tag *apk.Tag) bool {
      return tag.DecodedName == "application"
    })
    e := xml.Marshal(*manifest)
    if e != nil {
      panic(e)
    }

    debuggableXml, e := apk.DecodeXml(*manifest)
    if e != nil {
      panic(e)
    }
    if "debuggable" != debuggableXml.StrPool.Strs[debuggableXml.StrPool.StrCount-1] {
      panic("validate failed(no \"debuggable\" string found in pool)")
    }
    b := false
    for _, t := range debuggableXml.Tags {
      if t.DecodedName == "application" {
        for _, attr := range t.Attrs {
          if attr.DecodedFull == "android:debuggable=\"true\"" {
            b = true
            break
          }
        }
        break
      }
    }
    if !b {
      panic("validate failed(no \"debuggable\" attr found in application node)")
    }
  }

  if *json {
    e := xml.MarshalJSON(*manifest + ".json")
    if e != nil {
      panic(e)
    }
  }
}
