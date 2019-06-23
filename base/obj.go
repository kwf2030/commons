package base

import (
  "math/rand"

  "github.com/kwf2030/commons/time2"
)

var R = rand.New(rand.NewSource(time2.Timestamp()))

type Equality interface {
  Equals(interface{}) bool
}

type Comparable interface {
  CompareTo(interface{}) int
}

type Cloneable interface {
  Clone() interface{}
}
