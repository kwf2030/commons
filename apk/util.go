package apk

import (
  "bufio"
  "bytes"
  "io"
  "math"

  "github.com/kwf2030/commons/conv"
)

type Header struct {
  // 类型
  Type uint16

  // Header大小
  HeaderSize uint16

  // Chunk大小
  Size uint32
}

func parseHeader(r *bytesReader) *Header {
  return &Header{
    Type:       r.readUint16(),
    HeaderSize: r.readUint16(),
    Size:       r.readUint32(),
  }
}

func (h *Header) writeTo(w *bytesWriter) {
  w.writeUint16(h.Type)
  w.writeUint16(h.HeaderSize)
  w.writeUint32(h.Size)
}

type StrPool struct {
  // Chunk的起始和结束位置（非协议字段）
  ChunkStart, ChunkEnd uint32

  *Header

  // 字符串个数
  StrCount uint32

  // 字符串样式个数
  StyleCount uint32

  // 字符串标识，
  // SortedFlag: 0x0001
  // UTF16Flag:  0x0000
  // UTF8Flag:   0x0100
  Flags uint32

  // 字符串起始位置偏移（相对Header）
  StrStart uint32

  // 字符串样式起始位置偏移（相对Header）
  StyleStart uint32

  // 字符串偏移数组（相对Strs），长度为StrCount
  StrOffsets []uint32

  // 字符串样式偏移数组（相对Styles），长度为StyleCount
  StyleOffsets []uint32

  // 字符串，长度为StrCount，
  // 若是UTF-8编码，以0x00（1个字节）作为结束符，
  // 若是UTF-16编码，以0x0000（2个字节）作为结束符
  Strs []string

  Styles []byte
}

func parseStrPool(r *bytesReader) *StrPool {
  chunkStart := r.pos()
  header := parseHeader(r)
  strCount := r.readUint32()
  styleCount := r.readUint32()
  flags := r.readUint32()
  strStart := r.readUint32()
  styleStart := r.readUint32()
  strOffsets := r.readUint32Array(strCount)
  styleOffsets := r.readUint32Array(styleCount)

  var strs []string
  if strCount > 0 && styleCount < math.MaxUint32 {
    end := chunkStart + header.Size
    if styleCount > 0 && styleCount < math.MaxUint32 {
      end = chunkStart + styleStart
    }
    pool := r.slice(r.pos(), end)
    strs = make([]string, strCount)
    if flags&0x0100 != 0 {
      for i := uint32(0); i < strCount; i++ {
        strs[i] = string(str8(pool, strOffsets[i]))
      }
    } else {
      for i := uint32(0); i < strCount; i++ {
        strs[i] = string(str16(pool, strOffsets[i]))
      }
    }
  }

  // todo 样式解析
  var styles []byte
  if styleCount > 0 && styleCount < math.MaxUint32 {
    styles = r.slice(chunkStart+styleStart, chunkStart+header.Size)
  }

  return &StrPool{
    ChunkStart:   chunkStart,
    ChunkEnd:     chunkStart + header.Size,
    Header:       header,
    StrCount:     strCount,
    StyleCount:   styleCount,
    Flags:        flags,
    StrStart:     strStart,
    StyleStart:   styleStart,
    StrOffsets:   strOffsets,
    StyleOffsets: styleOffsets,
    Strs:         strs,
    Styles:       styles,
  }
}

func (p *StrPool) writeTo(w *bytesWriter) {
  p.Header.writeTo(w)
  w.writeUint32(p.StrCount)
  w.writeUint32(p.StyleCount)
  w.writeUint32(p.Flags)
  w.writeUint32(p.StrStart)
  w.writeUint32(p.StyleStart)
  w.writeUint32Array(p.StrOffsets)
  w.writeUint32Array(p.StyleOffsets)
  if len(p.Strs) > 0 {
    p.writeStrs(w)
  }
  if len(p.Styles) > 0 {
    w.Write(p.Styles)
  }
}

func (p *StrPool) writeStrs(w *bytesWriter) {
  if p.Flags&0x0100 != 0 {
    for _, str := range p.Strs {
      w.writeUint16(uint16(len(str)))
      w.Write([]byte(str))
      w.writeUint8(0)
    }
  } else {
    for _, str := range p.Strs {
      l := len(str)
      w.writeUint16(uint16(l))
      for i := 0; i < l; i++ {
        w.writeUint8(str[i])
        w.writeUint8(0)
      }
      w.writeUint16(0)
    }
  }
}

func str8(block []byte, offset uint32) []byte {
  n := 1
  if x := block[offset] & 0x80; x != 0 {
    n = 2
  }
  s := offset + uint32(n)
  b := block[s]
  if b == 0 {
    return nil
  }
  s++
  if b&0x80 != 0 {
    b = (b&0x7F)<<8 | block[s]&0xFF
    s++
  }
  return block[s : s+uint32(b)]
}

func str16(block []byte, offset uint32) []byte {
  // 2个字节表示字符串长度（去掉多余的0和结束符后的长度）
  n := 2
  // 如果第2个字节&0x10000000不为0，则是4个字节表示字符串长度
  if x := block[offset+1] & 0x80; x != 0 {
    n = 4
  }
  // 跳过长度
  s := offset + uint32(n)
  e := s
  l := uint32(len(block))
  for {
    if e+1 >= l {
      break
    }
    // 0x0000（连续2个字节是0）表示字符串结束
    if block[e] == 0 && block[e+1] == 0 {
      break
    }
    e += 2
  }
  // 去掉多余的0
  ret := make([]byte, 0, (e-s)/2)
  for _, v := range block[s:e] {
    if v != 0 {
      ret = append(ret, v)
    }
  }
  return ret
}

type bytesReader struct {
  *bytes.Reader
  data []byte
}

func newBytesReader(data []byte) *bytesReader {
  return &bytesReader{Reader: bytes.NewReader(data), data: data}
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

func (w *bytesWriter) writeUint8(n uint8) {
  w.WriteByte(n)
}

func (w *bytesWriter) writeUint16(n uint16) {
  w.Write(conv.Uint16ToBytesL(n))
}

func (w *bytesWriter) writeUint32(n uint32) {
  w.Write(conv.Uint32ToBytesL(n))
}

func (w *bytesWriter) writeUint32Array(arr []uint32) {
  if len(arr) < 1 {
    return
  }
  for _, n := range arr {
    w.Write(conv.Uint32ToBytesL(n))
  }
}
