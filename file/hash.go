package file

import (
  "bytes"
  "crypto/md5"
  "crypto/sha1"
  "crypto/sha256"
  "encoding/hex"
  "hash"
  "io"
  "os"
)

func MD5(path string) (string, error) {
  return Hash(path, md5.New())
}

func BytesMD5(data []byte) (string, error) {
  return BytesHash(data, md5.New())
}

func SHA1(path string) (string, error) {
  return Hash(path, sha1.New())
}

func BytesSHA1(data []byte) (string, error) {
  return BytesHash(data, sha1.New())
}

func SHA256(path string) (string, error) {
  return Hash(path, sha256.New())
}

func BytesSHA256(data []byte) (string, error) {
  return BytesHash(data, sha256.New())
}

func Hash(path string, hash hash.Hash) (string, error) {
  f, e := os.Open(path)
  if e != nil {
    return "", e
  }
  defer f.Close()
  if _, e = io.Copy(hash, f); e != nil {
    return "", e
  }
  return hex.EncodeToString(hash.Sum(nil)), nil
}

func BytesHash(data []byte, hash hash.Hash) (string, error) {
  if _, e := io.Copy(hash, bytes.NewReader(data)); e != nil {
    return "", e
  }
  return hex.EncodeToString(hash.Sum(nil)), nil
}
