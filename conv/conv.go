package conv

import (
  "strconv"
  "strings"
)

func Bool(any interface{}) bool {
  switch val := any.(type) {
  case bool:
    return val
  case string:
    return val != "" && strings.ToLower(val) != "false"
  case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
    return val != 0
  }
  return any != nil
}

func Int(any interface{}, defaultValue int) int {
  switch val := any.(type) {
  case int:
    return val
  case string:
    i, e := strconv.Atoi(val)
    if e == nil {
      return i
    }
  case float64:
    return int(val)
  case float32:
    return int(val)
  case int64:
    return int(val)
  case int32:
    return int(val)
  case int16:
    return int(val)
  case int8:
    return int(val)
  case uint:
    return int(val)
  case uint64:
    return int(val)
  case uint32:
    return int(val)
  case uint16:
    return int(val)
  case uint8:
    return int(val)
  case bool:
    if val {
      return 1
    } else {
      return 0
    }
  }
  return defaultValue
}

func Int64(any interface{}, defaultValue int64) int64 {
  switch val := any.(type) {
  case int64:
    return val
  case string:
    i, e := strconv.ParseInt(val, 10, 64)
    if e == nil {
      return i
    }
  case float64:
    return int64(val)
  case float32:
    return int64(val)
  case int:
    return int64(val)
  case int32:
    return int64(val)
  case int16:
    return int64(val)
  case int8:
    return int64(val)
  case uint:
    return int64(val)
  case uint64:
    return int64(val)
  case uint32:
    return int64(val)
  case uint16:
    return int64(val)
  case uint8:
    return int64(val)
  case bool:
    if val {
      return 1
    } else {
      return 0
    }
  }
  return defaultValue
}

func Uint(any interface{}, defaultValue uint) uint {
  switch val := any.(type) {
  case uint:
    return val
  case string:
    i, e := strconv.ParseUint(val, 10, 0)
    if e == nil {
      return uint(i)
    }
  case float64:
    return uint(val)
  case float32:
    return uint(val)
  case int:
    return uint(val)
  case int64:
    return uint(val)
  case int32:
    return uint(val)
  case int16:
    return uint(val)
  case int8:
    return uint(val)
  case uint64:
    return uint(val)
  case uint32:
    return uint(val)
  case uint16:
    return uint(val)
  case uint8:
    return uint(val)
  case bool:
    if val {
      return 1
    } else {
      return 0
    }
  }
  return defaultValue
}

func Uint64(any interface{}, defaultValue uint64) uint64 {
  switch val := any.(type) {
  case uint64:
    return val
  case string:
    i, e := strconv.ParseUint(val, 10, 64)
    if e == nil {
      return i
    }
  case float64:
    return uint64(val)
  case float32:
    return uint64(val)
  case int:
    return uint64(val)
  case int64:
    return uint64(val)
  case int32:
    return uint64(val)
  case int16:
    return uint64(val)
  case int8:
    return uint64(val)
  case uint:
    return uint64(val)
  case uint32:
    return uint64(val)
  case uint16:
    return uint64(val)
  case uint8:
    return uint64(val)
  case bool:
    if val {
      return 1
    } else {
      return 0
    }
  }
  return defaultValue
}
func String(any interface{}, defaultValue string) string {
  switch val := any.(type) {
  case string:
    return val
  case float64:
    return strconv.FormatFloat(val, 'f', 2, 64)
  case float32:
    return strconv.FormatFloat(float64(val), 'f', 2, 32)
  case int:
    return strconv.FormatInt(int64(val), 10)
  case int64:
    return strconv.FormatInt(val, 10)
  case int32:
    return strconv.FormatInt(int64(val), 10)
  case int16:
    return strconv.FormatInt(int64(val), 10)
  case int8:
    return strconv.FormatInt(int64(val), 10)
  case uint:
    return strconv.FormatUint(uint64(val), 10)
  case uint64:
    return strconv.FormatUint(val, 10)
  case uint32:
    return strconv.FormatUint(uint64(val), 10)
  case uint16:
    return strconv.FormatUint(uint64(val), 10)
  case uint8:
    return strconv.FormatUint(uint64(val), 10)
  case bool:
    if val {
      return "true"
    } else {
      return ""
    }
  case []byte:
    return string(val)
  }
  return defaultValue
}
