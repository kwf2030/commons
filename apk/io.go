package main

import (
  "bufio"
  "bytes"
  "io"

  "github.com/kwf2030/commons/conv"
)

type bytesReader struct {
  *bytes.Reader
  data []byte
}

func (r *bytesReader) pos() uint32 {
  return uint32(len(r.data) - r.Len())
}

func (r *bytesReader) slice(start, end uint32) []byte {
  r.Seek(int64(end), io.SeekStart)
  return r.data[start:end]
}

func (r *bytesReader) readN(n uint32) []byte {
  if n < 1 {
    return nil
  }
  ret := make([]byte, n)
  r.Read(ret)
  return ret
}

func (r *bytesReader) unreadN(n int64) {
  if n < 1 {
    return
  }
  r.Seek(-n, io.SeekCurrent)
}

func (r *bytesReader) readUint8() uint8 {
  b, _ := r.ReadByte()
  return uint8(b)
}

func (r *bytesReader) readUint16() uint16 {
  return conv.BytesToUint16L(r.readN(2))
}

func (r *bytesReader) readUint32() uint32 {
  return conv.BytesToUint32L(r.readN(4))
}

func (r *bytesReader) readUint32Array(count uint32) []uint32 {
  if count < 1 {
    return nil
  }
  ret := make([]uint32, count)
  for i := uint32(0); i < count; i++ {
    ret[i] = r.readUint32()
  }
  return ret
}

type bytesWriter struct {
  *bufio.Writer
}

func newBytesWriter(w io.Writer) *bytesWriter {
  return &bytesWriter{Writer: bufio.NewWriterSize(w, 1024*1024)}
}

func (w *bytesWriter) writeUint8(data uint8) {
  w.WriteByte(data)
}

func (w *bytesWriter) writeUint16(data uint16) {
  w.Write(conv.Uint16ToBytesL(data))
}

func (w *bytesWriter) writeUint32(data uint32) {
  w.Write(conv.Uint32ToBytesL(data))
}

func (w *bytesWriter) writeUint32Array(data []uint32) {
  if len(data) < 1 {
    return
  }
  for _, v := range data {
    w.Write(conv.Uint32ToBytesL(v))
  }
}
