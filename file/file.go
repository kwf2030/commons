package file

import "os"

func Exist(path string) bool {
  if path == "" {
    return false
  }
  _, e := os.Stat(path)
  if e != nil {
    return false
  }
  return true
}

func IsFile(path string) bool {
  if path == "" {
    return false
  }
  f, e := os.Stat(path)
  if e != nil {
    return false
  }
  return !f.IsDir()
}

func IsDir(path string) bool {
  if path == "" {
    return false
  }
  f, e := os.Stat(path)
  if e != nil {
    return false
  }
  return f.IsDir()
}
