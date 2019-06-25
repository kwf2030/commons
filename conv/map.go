package conv

func GetBool(m map[string]interface{}, key string, defaultValue bool) bool {
  if m == nil || key == "" {
    return defaultValue
  }
  if v, ok := m[key]; ok {
    return Bool(v)
  }
  return defaultValue
}

func GetInt(m map[string]interface{}, key string, defaultValue int) int {
  if m == nil || key == "" {
    return defaultValue
  }
  if v, ok := m[key]; ok {
    return Int(v, defaultValue)
  }
  return defaultValue
}

func GetInt64(m map[string]interface{}, key string, defaultValue int64) int64 {
  if m == nil || key == "" {
    return defaultValue
  }
  if v, ok := m[key]; ok {
    return Int64(v, defaultValue)
  }
  return defaultValue
}

func GetUint(m map[string]interface{}, key string, defaultValue uint) uint {
  if m == nil || key == "" {
    return defaultValue
  }
  if v, ok := m[key]; ok {
    return Uint(v, defaultValue)
  }
  return defaultValue
}

func GetUint64(m map[string]interface{}, key string, defaultValue uint64) uint64 {
  if m == nil || key == "" {
    return defaultValue
  }
  if v, ok := m[key]; ok {
    return Uint64(v, defaultValue)
  }
  return defaultValue
}

func GetString(m map[string]interface{}, key string, defaultValue string) string {
  if m == nil || key == "" {
    return defaultValue
  }
  if v, ok := m[key]; ok {
    return String(v, defaultValue)
  }
  return defaultValue
}

func GetMap(m map[string]interface{}, key string) map[string]interface{} {
  if m == nil || key == "" {
    return nil
  }
  if v, ok := m[key]; ok {
    if ret, ok := v.(map[string]interface{}); ok {
      return ret
    }
  }
  return nil
}

func GetMapSlice(m map[string]interface{}, key string) []map[string]interface{} {
  if m == nil || key == "" {
    return nil
  }
  if val, ok := m[key]; ok {
    switch v := val.(type) {
    case []interface{}:
      ret := make([]map[string]interface{}, 0, len(v))
      for _, vv := range v {
        if mm, ok := vv.(map[string]interface{}); ok {
          ret = append(ret, mm)
        }
      }
      return ret
    case []map[string]interface{}:
      return v
    }
  }
  return nil
}
