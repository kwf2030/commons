package conv

func GetBool(data map[string]interface{}, key string, value bool) bool {
  if data == nil || key == "" {
    return value
  }
  if v, ok := data[key]; ok {
    return Bool(v)
  }
  return value
}

func GetInt(data map[string]interface{}, key string, value int) int {
  if data == nil || key == "" {
    return value
  }
  if v, ok := data[key]; ok {
    return Int(v, value)
  }
  return value
}

func GetInt64(data map[string]interface{}, key string, value int64) int64 {
  if data == nil || key == "" {
    return value
  }
  if v, ok := data[key]; ok {
    return Int64(v, value)
  }
  return value
}

func GetUint(data map[string]interface{}, key string, value uint) uint {
  if data == nil || key == "" {
    return value
  }
  if v, ok := data[key]; ok {
    return Uint(v, value)
  }
  return value
}

func GetUint64(data map[string]interface{}, key string, value uint64) uint64 {
  if data == nil || key == "" {
    return value
  }
  if v, ok := data[key]; ok {
    return Uint64(v, value)
  }
  return value
}

func GetString(data map[string]interface{}, key string, value string) string {
  if data == nil || key == "" {
    return value
  }
  if v, ok := data[key]; ok {
    return String(v, value)
  }
  return value
}

func GetMap(data map[string]interface{}, key string) map[string]interface{} {
  if data == nil || key == "" {
    return nil
  }
  if v, ok := data[key]; ok {
    if ret, ok := v.(map[string]interface{}); ok {
      return ret
    }
  }
  return nil
}

func GetMapSlice(data map[string]interface{}, key string) []map[string]interface{} {
  if data == nil || key == "" {
    return nil
  }
  if v, ok := data[key]; ok {
    switch ret := v.(type) {
    case []interface{}:
      arr := make([]map[string]interface{}, 0, len(ret))
      for _, m := range ret {
        if vv, ok := m.(map[string]interface{}); ok {
          arr = append(arr, vv)
        }
      }
      return arr
    case []map[string]interface{}:
      return ret
    }
  }
  return nil
}
