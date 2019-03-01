package conv

import (
  "bytes"
  "io"
)

type BytesReader struct {
  *bytes.Reader
  data []byte
  l    bool
}

func NewBytesReader(data []byte) *BytesReader {
  return &BytesReader{Reader: bytes.NewReader(data), data: data}
}

func NewBytesReaderL(data []byte) *BytesReader {
  return &BytesReader{Reader: bytes.NewReader(data), data: data, l: true}
}

func (r *BytesReader) Pos() int {
  return len(r.data) - r.Len()
}

func (r *BytesReader) Slice(start, end int) []byte {
  r.Seek(int64(end), io.SeekStart)
  return r.data[start:end]
}

func (r *BytesReader) SliceRel(n int) []byte {
  r.Seek(int64(n), io.SeekCurrent)
  s := r.Pos()
  return r.data[s : s+n]
}

func (r *BytesReader) ReadN(n int) []byte {
  if n < 1 {
    return nil
  }
  ret := make([]byte, n)
  r.Read(ret)
  return ret
}

func (r *BytesReader) UnreadN(n int) {
  if n < 1 {
    return
  }
  r.Seek(int64(-n), io.SeekCurrent)
}

func (r *BytesReader) ReadUint8() uint8 {
  b, _ := r.ReadByte()
  return uint8(b)
}

func (r *BytesReader) ReadUint16() uint16 {
  b := r.ReadN(2)
  if r.l {
    return BytesToUint16L(b)
  }
  return BytesToUint16(b)
}

func (r *BytesReader) ReadUint32() uint32 {
  b := r.ReadN(4)
  if r.l {
    return BytesToUint32L(b)
  }
  return BytesToUint32(b)
}

func (r *BytesReader) readUint32Array(n int) []uint32 {
  if n < 1 {
    return nil
  }
  ret := make([]uint32, n)
  for i := 0; i < n; i++ {
    ret[i] = r.ReadUint32()
  }
  return ret
}
