package time2

import (
  "math/rand"
  "time"
)

const (
  DateFormat  = "2006-01-02"
  DateFormat2 = "2006_01_02"
  DateFormat3 = "2006/01/02"
  DateFormat4 = "2006.01.02"

  TimeFormat   = "15:04"
  TimeSFormat  = "15:04:05"
  TimeMsFormat = "15:04:05.000"

  DateTimeFormat  = "2006-01-02 15:04"
  DateTimeFormat2 = "2006_01_02 15:04"
  DateTimeFormat3 = "2006/01/02 15:04"
  DateTimeFormat4 = "2006.01.02 15:04"
  DateTimeFormat5 = "200601021504"

  DateTimeSFormat  = "2006-01-02 15:04:05"
  DateTimeSFormat2 = "2006_01_02 15:04:05"
  DateTimeSFormat3 = "2006/01/02 15:04:05"
  DateTimeSFormat4 = "2006.01.02 15:04:05"
  DateTimeSFormat5 = "20060102150405"

  DateTimeMsFormat  = "2006-01-02 15:04:05.000"
  DateTimeMsFormat2 = "2006_01_02 15:04:05.000"
  DateTimeMsFormat3 = "2006/01/02 15:04:05.000"
  DateTimeMsFormat4 = "2006.01.02 15:04:05.000"
  DateTimeMsFormat5 = "20060102150405000"
)

const (
  OneSecondInMillis    = 1000
  TwoSecondInMillis    = 2000
  ThreeSecondsInMillis = 3000
  FourSecondsInMillis  = 4000
  FiveSecondsInMillis  = 5000
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
  time.Sleep(RandMillis(OneSecondInMillis, FiveSecondsInMillis))
}

func UntilTomorrow() time.Duration {
  t1 := Now()
  t2 := t1.Add(time.Hour * 24)
  t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t1.Location())
  return t2.Sub(t1)
}
