package conv

import (
  "encoding/json"
  "io"
  "io/ioutil"

  "github.com/kwf2030/commons/base"
)

func MapToJson(m map[string]interface{}) ([]byte, error) {
  if len(m) == 0 {
    return nil, base.ErrInvalidArgument
  }
  ret, e := json.Marshal(m)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

func JsonToMap(data []byte) (map[string]interface{}, error) {
  if len(data) == 0 {
    return nil, base.ErrInvalidArgument
  }
  ret := make(map[string]interface{}, 16)
  e := json.Unmarshal(data, &ret)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

func ReadJson(r io.Reader, in interface{}) error {
  if r == nil || in == nil {
    return base.ErrInvalidArgument
  }
  data, e := ioutil.ReadAll(r)
  if e != nil {
    return e
  }
  return json.Unmarshal(data, in)
}

func ReadJsonToMap(r io.Reader) (map[string]interface{}, error) {
  if r == nil {
    return nil, base.ErrInvalidArgument
  }
  data, e := ioutil.ReadAll(r)
  if e != nil {
    return nil, e
  }
  return JsonToMap(data)
}
