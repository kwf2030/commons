package conv

import (
  "reflect"
  "unsafe"
)

func StrBytes(str string) []byte {
  var b reflect.SliceHeader
  s := (*reflect.StringHeader)(unsafe.Pointer(&str))
  b.Data, b.Len, b.Cap = s.Data, s.Len, s.Len
  return *(*[]byte)(unsafe.Pointer(&b))
}

func BytesStr(bytes []byte) string {
  return *(*string)(unsafe.Pointer(&bytes))
}
