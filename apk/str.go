package main

func str8(data []byte, offset uint32) []byte {
  n := 1
  if x := data[offset] & 0x80; x != 0 {
    n = 2
  }
  s := offset + uint32(n)
  b := data[s]
  if b == 0 {
    return nil
  }
  s++
  if b&0x80 != 0 {
    b = (b&0x7F)<<8 | data[s]&0xFF
    s++
  }
  return data[s : s+uint32(b)]
}

func str16(data []byte, offset uint32) []byte {
  // 2个字节表示字符串长度（去掉多余空格和结束符的长度）
  n := 2
  // 如果第2个字节&0x10000000不为0，则是4个字节表示字符串长度
  if x := data[offset+1] & 0x80; x != 0 {
    n = 4
  }
  // 跳过长度
  s := offset + uint32(n)
  e := s
  l := uint32(len(data))
  for {
    if e+1 >= l {
      break
    }
    // 0x0000（连续2个字节是0）表示字符串结束
    if data[e] == 0 && data[e+1] == 0 {
      break
    }
    e += 2
  }
  // 去掉多余的0
  ret := make([]byte, 0, (e-s)/2)
  for _, v := range data[s:e] {
    if v != 0 {
      ret = append(ret, v)
    }
  }
  return ret
}
