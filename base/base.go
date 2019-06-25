package base

import (
  "errors"
  "math/rand"
  "time"
)

var (
  ErrNilPointer      = errors.New("nil pointer")
  ErrInvalidArgument = errors.New("invalid argument")
  ErrIndexOutOfRange = errors.New("index out of range")
  ErrTimeout         = errors.New("timeout")
)

var Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type Equality interface {
  Equals(interface{}) bool
}

type Comparable interface {
  CompareTo(interface{}) int
}

type Cloneable interface {
  Clone() interface{}
}
