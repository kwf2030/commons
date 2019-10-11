package base

import (
  "errors"
)

var (
  ErrNilPointer      = errors.New("nil pointer")
  ErrInvalidArgument = errors.New("invalid argument")
  ErrIndexOutOfRange = errors.New("index out of range")
  ErrTimeout         = errors.New("timeout")
)

type Equality interface {
  Equals(interface{}) bool
}

type Comparable interface {
  CompareTo(interface{}) int
}

type Cloneable interface {
  Clone() interface{}
}
