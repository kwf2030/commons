package base

import (
  "errors"
  "math/rand"
  "time"
)

var (
  ErrNullPointer     = errors.New("error: null pointer")
  ErrInvalidArgs     = errors.New("error: invalid args")
  ErrIndexOutOfRange = errors.New("error: index out of range")
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
