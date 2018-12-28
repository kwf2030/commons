package conv

func GetBool(data map[string]interface{}, key string, defaultValue bool) bool {
  if data == nil || key == "" {
    return defaultValue
  }
  if v, ok := data[key]; ok {
    return Bool(v)
  }
  return defaultValue
}

func GetInt(data map[string]interface{}, key string, defaultValue int) int {
  if data == nil || key == "" {
    return defaultValue
  }
  if v, ok := data[key]; ok {
    return Int(v, defaultValue)
  }
  return defaultValue
}

func GetInt64(data map[string]interface{}, key string, defaultValue int64) int64 {
  if data == nil || key == "" {
    return defaultValue
  }
  if v, ok := data[key]; ok {
    return Int64(v, defaultValue)
  }
  return defaultValue
}

func GetUint(data map[string]interface{}, key string, defaultValue uint) uint {
  if data == nil || key == "" {
    return defaultValue
  }
  if v, ok := data[key]; ok {
    return Uint(v, defaultValue)
  }
  return defaultValue
}

func GetUint64(data map[string]interface{}, key string, defaultValue uint64) uint64 {
  if data == nil || key == "" {
    return defaultValue
  }
  if v, ok := data[key]; ok {
    return Uint64(v, defaultValue)
  }
  return defaultValue
}

func GetString(data map[string]interface{}, key string, defaultValue string) string {
  if data == nil || key == "" {
    return defaultValue
  }
  if v, ok := data[key]; ok {
    return String(v, defaultValue)
  }
  return defaultValue
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
