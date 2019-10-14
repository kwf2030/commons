package conv

import (
  "reflect"
  "strconv"
  "unsafe"
)

func StrToBytes(str string) []byte {
  var b reflect.SliceHeader
  s := (*reflect.StringHeader)(unsafe.Pointer(&str))
  b.Data, b.Len, b.Cap = s.Data, s.Len, s.Len
  return *(*[]byte)(unsafe.Pointer(&b))
}

func BytesToStr(bytes []byte) string {
  return *(*string)(unsafe.Pointer(&bytes))
}

func NumberToBytes(n int) []byte {
  return StrToBytes(strconv.Itoa(n))
}

func BytesToNumber(bytes []byte) int {
  str := BytesToStr(bytes)
  n, _ := strconv.Atoi(str)
  return n
}
