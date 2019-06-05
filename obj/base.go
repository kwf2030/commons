package obj

type Equality interface {
  Equals(interface{}) bool

  HashCode() int
}

type Comparable interface {
  CompareTo(interface{}) int
}

type Cloneable interface {
  Clone() interface{}
}
