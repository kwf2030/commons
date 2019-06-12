package conv

import (
  "strconv"
  "strings"
)

func Bool(any interface{}) bool {
  switch ret := any.(type) {
  case bool:
    return ret
  case string:
    return ret != "" && strings.ToLower(ret) != "false"
  case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
    return ret != 0
  }
  return any != nil
}

func Int(any interface{}, defaultValue int) int {
  switch ret := any.(type) {
  case int:
    return ret
  case string:
    i, e := strconv.Atoi(ret)
    if e == nil {
      return i
    }
  case float64:
    return int(ret)
  case float32:
    return int(ret)
  case int64:
    return int(ret)
  case int32:
    return int(ret)
  case int16:
    return int(ret)
  case int8:
    return int(ret)
  case uint:
    return int(ret)
  case uint64:
    return int(ret)
  case uint32:
    return int(ret)
  case uint16:
    return int(ret)
  case uint8:
    return int(ret)
  case bool:
    if ret {
      return 1
    } else {
      return 0
    }
  }
  return defaultValue
}

func Int64(any interface{}, defaultValue int64) int64 {
  switch ret := any.(type) {
  case int64:
    return ret
  case string:
    i, e := strconv.ParseInt(ret, 10, 64)
    if e == nil {
      return i
    }
  case float64:
    return int64(ret)
  case float32:
    return int64(ret)
  case int:
    return int64(ret)
  case int32:
    return int64(ret)
  case int16:
    return int64(ret)
  case int8:
    return int64(ret)
  case uint:
    return int64(ret)
  case uint64:
    return int64(ret)
  case uint32:
    return int64(ret)
  case uint16:
    return int64(ret)
  case uint8:
    return int64(ret)
  case bool:
    if ret {
      return 1
    } else {
      return 0
    }
  }
  return defaultValue
}

func Uint(any interface{}, defaultValue uint) uint {
  switch ret := any.(type) {
  case uint:
    return ret
  case string:
    i, e := strconv.ParseUint(ret, 10, 0)
    if e == nil {
      return uint(i)
    }
  case float64:
    return uint(ret)
  case float32:
    return uint(ret)
  case int:
    return uint(ret)
  case int64:
    return uint(ret)
  case int32:
    return uint(ret)
  case int16:
    return uint(ret)
  case int8:
    return uint(ret)
  case uint64:
    return uint(ret)
  case uint32:
    return uint(ret)
  case uint16:
    return uint(ret)
  case uint8:
    return uint(ret)
  case bool:
    if ret {
      return 1
    } else {
      return 0
    }
  }
  return defaultValue
}

func Uint64(any interface{}, defaultValue uint64) uint64 {
  switch ret := any.(type) {
  case uint64:
    return ret
  case string:
    i, e := strconv.ParseUint(ret, 10, 64)
    if e == nil {
      return i
    }
  case float64:
    return uint64(ret)
  case float32:
    return uint64(ret)
  case int:
    return uint64(ret)
  case int64:
    return uint64(ret)
  case int32:
    return uint64(ret)
  case int16:
    return uint64(ret)
  case int8:
    return uint64(ret)
  case uint:
    return uint64(ret)
  case uint32:
    return uint64(ret)
  case uint16:
    return uint64(ret)
  case uint8:
    return uint64(ret)
  case bool:
    if ret {
      return 1
    } else {
      return 0
    }
  }
  return defaultValue
}
func String(any interface{}, defaultValue string) string {
  switch ret := any.(type) {
  case string:
    return ret
  case float64:
    return strconv.FormatFloat(ret, 'f', 2, 64)
  case float32:
    return strconv.FormatFloat(float64(ret), 'f', 2, 32)
  case int:
    return strconv.FormatInt(int64(ret), 10)
  case int64:
    return strconv.FormatInt(ret, 10)
  case int32:
    return strconv.FormatInt(int64(ret), 10)
  case int16:
    return strconv.FormatInt(int64(ret), 10)
  case int8:
    return strconv.FormatInt(int64(ret), 10)
  case uint:
    return strconv.FormatUint(uint64(ret), 10)
  case uint64:
    return strconv.FormatUint(ret, 10)
  case uint32:
    return strconv.FormatUint(uint64(ret), 10)
  case uint16:
    return strconv.FormatUint(uint64(ret), 10)
  case uint8:
    return strconv.FormatUint(uint64(ret), 10)
  case bool:
    if ret {
      return "true"
    } else {
      return ""
    }
  }
  return defaultValue
}
