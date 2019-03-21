package main

import (
  "flag"
  "fmt"

  "github.com/kwf2030/commons/apk"
)

func main() {
  manifest := flag.String("m", "AndroidManifest.xml", "specify manifest file")
  debuggable := flag.Bool("d", false, "set debuggable=\"true\" for application node")
  json := flag.Bool("j", false, "generate json file for manifest")
  flag.Parse()

  xml2 := apk.NewXml2(apk.ParseXml(*manifest))
  if xml2 == nil {
    fmt.Println("can not read manifest file")
    return
  }
  if *debuggable {
    xml2.AddAttr("android:debuggable", true, func(tag2 *apk.XmlTag2) bool {
      return tag2.Name == "application"
    })
    e := xml2.WriteToFile(*manifest)
    if e != nil {
      panic(e)
    }

    debuggableXml2 := apk.NewXml2(apk.ParseXml(*manifest))
    if "debuggable" != debuggableXml2.Ori.StrPool.Strs[debuggableXml2.Ori.StrPool.StrCount-1] {
      panic("validate failed(no \"debuggable\" string found in pool)")
    }
    b := false
    for _, t := range debuggableXml2.Tags2 {
      if t.Name == "application" {
        for _, attr := range t.Attrs {
          if attr == "android:debuggable=\"true\"" {
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
    e := xml2.WriteToJsonFile(*manifest + ".json")
    if e != nil {
      panic(e)
    }
  }
}
