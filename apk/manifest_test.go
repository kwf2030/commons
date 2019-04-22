package apk

import (
  "os"
  "path"
  "testing"
)

func TestManifestModify(t *testing.T) {
  name := path.Join("testdata", "AndroidManifest")
  m1, _ := DecodeXml(name + ".xml")
  m1.MarshalJSON(name + ".json")
  m1.AddAttr("android:debuggable", true, 3, 4, 0, func(tag *Tag) bool {
    return tag.DecodedName == "application"
  })
  m1.AddResId(16842767, 4)
  m1.Marshal(name + "2.xml")
  m1.MarshalJSON(name + "2.json")
  // os.Remove(name + "2.xml")
  // os.Remove(name + "2.json")
}

func TestManifestRestore(t *testing.T) {
  name := path.Join("testdata", "AndroidManifest")
  m1, _ := DecodeXml(name + ".xml")
  m1.Marshal(name + "2.xml")
  m2, _ := DecodeXml(name + "2.xml")
  assertHeaderEquals(t, m1.Header, m2.Header)
  assertStrPoolEquals(t, m1.StrPool, m2.StrPool)
  assertResIdEquals(t, m1.ResId, m2.ResId)
  for i := 0; i < len(m1.Namespaces); i++ {
    assertNamespaceEquals(t, m1.Namespaces[i], m2.Namespaces[i])
  }
  for i := 0; i < len(m1.Tags); i++ {
    assertTagEquals(t, m1.Tags[i], m2.Tags[i])
  }
  os.Remove(name + "2.xml")
}

func assertHeaderEquals(t *testing.T, h1, h2 *Header) {
  assertUint16Equals(t, h1.Type, h2.Type)
  assertUint16Equals(t, h1.HeaderSize, h2.HeaderSize)
  assertUint32Equals(t, h1.Size, h2.Size)
}

func assertStrPoolEquals(t *testing.T, p1, p2 *StrPool) {
  assertHeaderEquals(t, p1.Header, p2.Header)
  assertUint32Equals(t, p1.StrCount, p2.StrCount)
  assertUint32Equals(t, p1.StyleCount, p2.StyleCount)
  assertUint32Equals(t, p1.Flags, p2.Flags)
  assertUint32Equals(t, p1.StrStart, p2.StrStart)
  assertUint32Equals(t, p1.StyleStart, p2.StyleStart)
  assertUint32ArrayEquals(t, p1.StrOffsets, p2.StrOffsets)
  assertUint32ArrayEquals(t, p1.StyleOffsets, p2.StyleOffsets)
  assertStrArrayEquals(t, p1.Strs, p2.Strs)
  assertByteArrayEquals(t, p1.Styles, p2.Styles)
}

func assertResIdEquals(t *testing.T, r1, r2 *ResId) {
  assertHeaderEquals(t, r1.Header, r2.Header)
  assertUint32ArrayEquals(t, r1.Ids, r2.Ids)
}

func assertNamespaceEquals(t *testing.T, ns1, ns2 *Namespace) {
  assertHeaderEquals(t, ns1.Header, ns2.Header)
  assertUint32Equals(t, ns1.LineNumber, ns2.LineNumber)
  assertUint32Equals(t, ns1.Res0, ns2.Res0)
  assertUint32Equals(t, ns1.Prefix, ns2.Prefix)
  assertUint32Equals(t, ns1.Uri, ns2.Uri)
}

func assertTagEquals(t *testing.T, t1, t2 *Tag) {
  assertHeaderEquals(t, t1.Header, t2.Header)
  assertUint32Equals(t, t1.LineNumber, t2.LineNumber)
  assertUint32Equals(t, t1.Res0, t2.Res0)
  assertUint32Equals(t, t1.NamespaceUri, t2.NamespaceUri)
  assertUint32Equals(t, t1.Name, t2.Name)
  assertUint16Equals(t, t1.AttrStart, t2.AttrStart)
  assertUint16Equals(t, t1.AttrSize, t2.AttrSize)
  assertUint16Equals(t, t1.AttrCount, t2.AttrCount)
  assertUint16Equals(t, t1.IdIndex, t2.IdIndex)
  assertUint16Equals(t, t1.ClassIndex, t2.ClassIndex)
  assertUint16Equals(t, t1.StyleIndex, t2.StyleIndex)
  for i := uint16(0); i < t1.AttrCount; i++ {
    assertAttrEquals(t, t1.Attrs[i], t2.Attrs[i])
  }
}

func assertAttrEquals(t *testing.T, a1, a2 *Attr) {
  assertUint32Equals(t, a1.NamespaceUri, a2.NamespaceUri)
  assertUint32Equals(t, a1.Name, a2.Name)
  assertUint32Equals(t, a1.RawValue, a2.RawValue)
  assertUint16Equals(t, a1.ValueSize, a2.ValueSize)
  assertUint8Equals(t, a1.Res0, a2.Res0)
  assertUint8Equals(t, a1.DataType, a2.DataType)
  assertUint32Equals(t, a1.Data, a2.Data)
}

func assertStrEquals(t *testing.T, str1, str2 string) {
  if str1 != str2 {
    t.Fatalf("assertStrEquals failed, str1=%s, str2=%s", str1, str2)
  }
}

func assertUint8Equals(t *testing.T, n1, n2 uint8) {
  if n1 != n2 {
    t.Fatalf("assertUint8Equals failed, n1=%d, n2=%d", n1, n2)
  }
}

func assertUint16Equals(t *testing.T, n1, n2 uint16) {
  if n1 != n2 {
    t.Fatalf("assertUint16Equals failed, n1=%d, n2=%d", n1, n2)
  }
}

func assertUint32Equals(t *testing.T, n1, n2 uint32) {
  if n1 != n2 {
    t.Fatalf("assertUint32Equals failed, n1=%d, n2=%d", n1, n2)
  }
}

func assertByteArrayEquals(t *testing.T, arr1, arr2 []byte) {
  l1, l2 := len(arr1), len(arr2)
  if l1 != l2 {
    t.Fatalf("assertByteArrayEquals failed, len(arr1)=%d, len(arr2)=%d", l1, l2)
    return
  }
  for i := 0; i < l1; i++ {
    if arr1[i] != arr2[i] {
      t.Fatalf("assertByteArrayEquals failed, arr1[%d]=%d, arr2[%d]=%d", i, arr1[i], i, arr2[i])
    }
  }
}

func assertStrArrayEquals(t *testing.T, arr1, arr2 []string) {
  l1, l2 := len(arr1), len(arr2)
  if l1 != l2 {
    t.Fatalf("assertStrArrayEquals failed, len(arr1)=%d, len(arr2)=%d", l1, l2)
    return
  }
  for i := 0; i < l1; i++ {
    if arr1[i] != arr2[i] {
      t.Fatalf("assertStrArrayEquals failed, arr1[%d]=%s, arr2[%d]=%s", i, arr1[i], i, arr2[i])
    }
  }
}

func assertUint32ArrayEquals(t *testing.T, arr1, arr2 []uint32) {
  l1, l2 := len(arr1), len(arr2)
  if l1 != l2 {
    t.Fatalf("assertUint32ArrayEquals failed, len(arr1)=%d, len(arr2)=%d", l1, l2)
    return
  }
  for i := 0; i < l1; i++ {
    if arr1[i] != arr2[i] {
      t.Fatalf("assertUint32ArrayEquals failed, arr1[%d]=%d, arr2[%d]=%d", i, arr1[i], i, arr2[i])
    }
  }
}
