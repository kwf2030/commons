package conv

import (
  "encoding/binary"
)

func Uint64ToBytes(i uint64) []byte {
  b := make([]byte, 8)
  binary.BigEndian.PutUint64(b, i)
  return b
}

func Uint64ToBytesL(i uint64) []byte {
  b := make([]byte, 8)
  binary.LittleEndian.PutUint64(b, i)
  return b
}

func Uint32ToBytes(i uint32) []byte {
  b := make([]byte, 4)
  binary.BigEndian.PutUint32(b, i)
  return b
}

func Uint32ToBytesL(i uint32) []byte {
  b := make([]byte, 4)
  binary.LittleEndian.PutUint32(b, i)
  return b
}

func Uint16ToBytes(i uint16) []byte {
  b := make([]byte, 2)
  binary.BigEndian.PutUint16(b, i)
  return b
}

func Uint16ToBytesL(i uint16) []byte {
  b := make([]byte, 2)
  binary.LittleEndian.PutUint16(b, i)
  return b
}

func BytesToUint64(data []byte) uint64 {
  return binary.BigEndian.Uint64(data)
}

func BytesToUint64L(data []byte) uint64 {
  return binary.LittleEndian.Uint64(data)
}

func BytesToUint32(data []byte) uint32 {
  return binary.BigEndian.Uint32(data)
}

func BytesToUint32L(data []byte) uint32 {
  return binary.LittleEndian.Uint32(data)
}

func BytesToUint16(data []byte) uint16 {
  return binary.BigEndian.Uint16(data)
}

func BytesToUint16L(data []byte) uint16 {
  return binary.LittleEndian.Uint16(data)
}
