package rand2

import (
  "math/rand"
  "time"
)

var (
  R = rand.New(rand.NewSource(time.Now().UnixNano()))

  numbers = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

  chars = []byte{
    '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
    'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
    'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
  }
)

func RandStr(length int) string {
  if length <= 0 {
    return ""
  }
  l := len(chars)
  bytes := make([]byte, length)
  for i := 0; i < length; i++ {
    bytes[i] = chars[R.Intn(l)]
  }
  return string(bytes)
}

func RandStrRange(minLen, maxLen int) string {
  if maxLen <= minLen {
    return RandStr(minLen)
  }
  return RandStr(R.Intn(maxLen+1-minLen) + minLen)
}

func RandStrFixed(lens []int) string {
  if len(lens) == 0 {
    return ""
  }
  return RandStr(lens[R.Intn(len(lens))])
}

func RandNumberStr(length int) string {
  if length <= 0 {
    return ""
  }
  l := len(numbers)
  bytes := make([]byte, length)
  for i := 0; i < length; i++ {
    bytes[i] = numbers[R.Intn(l)]
  }
  return string(bytes)
}

func RandNumberStrRange(minLen, maxLen int) string {
  if maxLen <= minLen {
    return RandNumberStr(minLen)
  }
  return RandNumberStr(R.Intn(maxLen+1-minLen) + minLen)
}

func RandNumberStrFixed(lens []int) string {
  if len(lens) == 0 {
    return ""
  }
  return RandNumberStr(lens[R.Intn(len(lens))])
}
