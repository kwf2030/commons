package conv

import (
  "bytes"
  "encoding/gob"
  "io"
  "io/ioutil"

  "github.com/kwf2030/commons/base"
)

func MapToGob(data map[string]interface{}) ([]byte, error) {
  if len(data) == 0 {
    return nil, base.ErrInvalidArgs
  }
  var buf bytes.Buffer
  e := gob.NewEncoder(&buf).Encode(data)
  if e != nil {
    return nil, e
  }
  return buf.Bytes(), nil
}

func GobToMap(data []byte) (map[string]interface{}, error) {
  if len(data) == 0 {
    return nil, base.ErrInvalidArgs
  }
  ret := make(map[string]interface{}, 16)
  e := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&ret)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

func ReadGob(r io.Reader, in interface{}) error {
  if r == nil || in == nil {
    return base.ErrInvalidArgs
  }
  data, e := ioutil.ReadAll(r)
  if e != nil {
    return e
  }
  return gob.NewDecoder(bytes.NewBuffer(data)).Decode(in)
}

func ReadGobToMap(r io.Reader) (map[string]interface{}, error) {
  if r == nil {
    return nil, base.ErrInvalidArgs
  }
  data, e := ioutil.ReadAll(r)
  if e != nil {
    return nil, e
  }
  return GobToMap(data)
}
