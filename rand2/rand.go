package rand2

import (
  cr "crypto/rand"
  "math/rand"
  "time"

  "github.com/kwf2030/commons/base"
)

var (
  numbers = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

  chars = []byte{
    '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
    'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
    'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
  }
)

func RandMilliseconds(min, max int) time.Duration {
  return time.Millisecond * time.Duration(rand.Intn(max+1-min)+min)
}

func RandStr(length int) string {
  if length <= 0 {
    return ""
  }
  l := len(chars)
  bytes := make([]byte, length)
  for i := 0; i < length; i++ {
    bytes[i] = chars[rand.Intn(l)]
  }
  return string(bytes)
}

func RandStrRange(minLen, maxLen int) string {
  if maxLen <= minLen {
    return RandStr(minLen)
  }
  return RandStr(rand.Intn(maxLen+1-minLen) + minLen)
}

func RandStrIn(lens []int) string {
  if len(lens) == 0 {
    return ""
  }
  return RandStr(lens[rand.Intn(len(lens))])
}

func RandNumberStr(length int) string {
  if length <= 0 {
    return ""
  }
  l := len(numbers)
  bytes := make([]byte, length)
  for i := 0; i < length; i++ {
    bytes[i] = numbers[rand.Intn(l)]
  }
  return string(bytes)
}

func RandNumberStrRange(minLen, maxLen int) string {
  if maxLen <= minLen {
    return RandNumberStr(minLen)
  }
  return RandNumberStr(rand.Intn(maxLen+1-minLen) + minLen)
}

func RandNumberStrIn(lens []int) string {
  if len(lens) == 0 {
    return ""
  }
  return RandNumberStr(lens[rand.Intn(len(lens))])
}

func RandCryptoBytes(length int) ([]byte, error) {
  if length <= 0 {
    return nil, base.ErrInvalidArgument
  }
  ret := make([]byte, length)
  _, e := cr.Read(ret)
  if e != nil {
    return nil, e
  }
  return ret, nil
}
