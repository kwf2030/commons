package base

import "strconv"

var (
  ErrNullPointer = NewException(-0xFF7801, "NullPointer Exception")
  ErrInvalidArgs = NewException(-0xFF7802, "InvalidArgs Exception")
)

type Exception struct {
  Err int
  Msg string
}

func NewException(err int, msg string) Exception {
  return Exception{err, msg}
}

func (e Exception) String() string {
  if e.Msg == "" {
    return "error<" + strconv.Itoa(e.Err) + ">"
  }
  return "error<" + strconv.Itoa(e.Err) + ">: " + e.Msg
}

func (e Exception) Error() string {
  return e.String()
}

func (e Exception) Equals(obj interface{}) bool {
  if err, ok := obj.(Exception); ok {
    return e.Err == err.Err && e.Msg == err.Msg
  }
  return false
}

func (e Exception) Clone() interface{} {
  return NewException(e.Err, e.Msg)
}
