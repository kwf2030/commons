package base

import (
  "math/rand"
  "time"
)

var R = rand.New(rand.NewSource(time.Now().UnixNano()))

type Equality interface {
  Equals(interface{}) bool
}

type Comparable interface {
  CompareTo(interface{}) int
}

type Cloneable interface {
  Clone() interface{}
}
