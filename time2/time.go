package time2

import (
  "time"

  "github.com/kwf2030/commons/base"
)

const (
  DateFormat  = "2006-01-02"
  DateFormat2 = "2006_01_02"
  DateFormat3 = "2006/01/02"
  DateFormat4 = "2006.01.02"

  TimeFormat   = "15:04"
  TimeFormatS  = "15:04:05"
  TimeFormatMs = "15:04:05.000"

  DateTimeFormat  = "2006-01-02 15:04"
  DateTimeFormat2 = "2006_01_02 15:04"
  DateTimeFormat3 = "2006/01/02 15:04"
  DateTimeFormat4 = "2006.01.02 15:04"
  DateTimeFormat5 = "200601021504"

  DateTimeFormatS  = "2006-01-02 15:04:05"
  DateTimeFormatS2 = "2006_01_02 15:04:05"
  DateTimeFormatS3 = "2006/01/02 15:04:05"
  DateTimeFormatS4 = "2006.01.02 15:04:05"
  DateTimeFormatS5 = "20060102150405"

  DateTimeFormatMs  = "2006-01-02 15:04:05.000"
  DateTimeFormatMs2 = "2006_01_02 15:04:05.000"
  DateTimeFormatMs3 = "2006/01/02 15:04:05.000"
  DateTimeFormatMs4 = "2006.01.02 15:04:05.000"
  DateTimeFormatMs5 = "20060102150405000"
)

var (
  TimeZoneSH, _ = time.LoadLocation("Asia/Shanghai")

  Nil time.Time
  // 0001-01-01 00:00:00
  NilStr = Nil.Format(DateTimeFormatS)
)

func Timestamp() int64 {
  return time.Now().Unix()
}

func TimestampNano() int64 {
  return time.Now().UnixNano()
}

func Now() time.Time {
  return time.Now().In(TimeZoneSH)
}

func NowStr() string {
  return Now().Format(DateTimeFormatS)
}

func NowStrf(format string) string {
  return Now().Format(format)
}

func Now2Tomorrow() time.Duration {
  t1 := Now()
  t2 := t1.Add(time.Hour * 24)
  t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t1.Location())
  return t2.Sub(t1)
}

func UTC() time.Time {
  return time.Now().UTC()
}

func UTCStr() string {
  return UTC().Format(DateTimeFormatS)
}

func UTCStrf(format string) string {
  return UTC().Format(format)
}

func UTC2Tomorrow() time.Duration {
  t1 := UTC()
  t2 := t1.Add(time.Hour * 24)
  t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t1.Location())
  return t2.Sub(t1)
}

func RandMillis(min, max int) time.Duration {
  n := base.R.Intn(max)
  if n < min {
    n = min
  }
  return time.Millisecond * time.Duration(n)
}
