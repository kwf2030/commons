package times

import (
  "math/rand"
  "time"
)

const (
  DateFormat       = "2006-01-02"
  DateTimeFormat   = "2006-01-02 15:04"
  DateTimeSFormat  = "2006-01-02 15:04:05"
  DateTimeMsFormat = "2006-01-02 15:04:05.000"

  DateFormat2       = "2006_01_02"
  DateTimeFormat2   = "2006_01_02 15:04"
  DateTimeSFormat2  = "2006_01_02 15:04:05"
  DateTimeMsFormat2 = "2006_01_02 15:04:05.000"

  DateFormat3       = "2006/01/02"
  DateTimeFormat3   = "2006/01/02 15:04"
  DateTimeSFormat3  = "2006/01/02 15:04:05"
  DateTimeMsFormat3 = "2006/01/02 15:04:05.000"

  DateFormat4       = "2006.01.02"
  DateTimeFormat4   = "2006.01.02 15:04"
  DateTimeSFormat4  = "2006.01.02 15:04:05"
  DateTimeMsFormat4 = "2006.01.02 15:04:05.000"

  DateFormat5       = "20060102"
  DateTimeFormat5   = "200601021504"
  DateTimeSFormat5  = "20060102150405"
  DateTimeMsFormat5 = "20060102150405000"
)

const (
  OneSecondInMillis    = 1000
  ThreeSecondsInMillis = 3000
)

var (
  TimeZoneSH, _ = time.LoadLocation("Asia/Shanghai")

  Empty    time.Time
  emptyStr string

  rnd = rand.New(rand.NewSource(Timestamp()))
)

func EmptyStr() string {
  if emptyStr == "" {
    emptyStr = Empty.Format(DateTimeSFormat)
  }
  return emptyStr
}

func Now() time.Time {
  return time.Now().In(TimeZoneSH)
}

func NowStr() string {
  return Now().Format(DateTimeSFormat)
}

func NowStrf(format string) string {
  return Now().Format(format)
}

func Timestamp() int64 {
  return Now().UnixNano()
}

func RandMillis(min, max int) time.Duration {
  n := rnd.Intn(max)
  if n < min {
    n = min
  }
  return time.Millisecond * time.Duration(n)
}

func Sleep() {
  time.Sleep(RandMillis(OneSecondInMillis, ThreeSecondsInMillis))
}

func UntilTomorrow() time.Duration {
  t1 := Now()
  t2 := t1.Add(time.Hour * 24)
  t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t1.Location())
  return t2.Sub(t1)
}
