package conv

import (
  "testing"
)

var str = "Go is expressive, concise, clean, and efficient. Its concurrency mechanisms make it easy to write programs that get the most out of multicore and networked machines, while its novel type system enables flexible and modular program construction. Go compiles quickly to machine code yet has the convenience of garbage collection and the power of run-time reflection. It's a fast, statically typed, compiled language that feels like a dynamically typed, interpreted language."

func BenchmarkStrToBytes1(b *testing.B) {
  for i := 0; i < b.N; i++ {
    _ = []byte(str)
  }
}

func BenchmarkStrToBytes2(b *testing.B) {
  for i := 0; i < b.N; i++ {
    _ = StrBytes(str)
  }
}

func BenchmarkBytesToStr1(b *testing.B) {
  bytes := []byte(str)
  for i := 0; i < b.N; i++ {
    _ = string(bytes)
  }
}

func BenchmarkBytesToStr2(b *testing.B) {
  bytes := []byte(str)
  for i := 0; i < b.N; i++ {
    _ = BytesStr(bytes)
  }
}
