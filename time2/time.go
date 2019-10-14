package time2

import (
  "time"
)

const (
  DateFormat  = "2006-01-02"
  DateFormat2 = "2006_01_02"
  DateFormat3 = "2006/01/02"
  DateFormat4 = "2006.01.02"

  TimeFormat    = "15:04"
  TimeFormatSec = "15:04:05"
  TimeFormatMs  = "15:04:05.000"

  DateTimeFormat    = "2006-01-02 15:04"
  DateTimeFormatSec = "2006-01-02 15:04:05"
  DateTimeFormatMs  = "2006-01-02 15:04:05.000"

  DateTimeFormat2    = "2006_01_02 15:04"
  DateTimeFormatSec2 = "2006_01_02 15:04:05"
  DateTimeFormatMs2  = "2006_01_02 15:04:05.000"

  DateTimeFormat3    = "2006/01/02 15:04"
  DateTimeFormatSec3 = "2006/01/02 15:04:05"
  DateTimeFormatMs3  = "2006/01/02 15:04:05.000"

  DateTimeFormat4    = "2006.01.02 15:04"
  DateTimeFormatSec4 = "2006.01.02 15:04:05"
  DateTimeFormatMs4  = "2006.01.02 15:04:05.000"

  DateTimeFormat5    = "200601021504"
  DateTimeFormatSec5 = "20060102150405"
  DateTimeFormatMs5  = "20060102150405000"
)

var (
  TimeZoneSH, _ = time.LoadLocation("Asia/Shanghai")

  Nil time.Time

  // 0001-01-01 00:00:00
  NilStr = Nil.Format(DateTimeFormatSec)
)

func Timestamp() int64 {
  return time.Now().Unix()
}

func TimestampNano() int64 {
  return time.Now().UnixNano()
}

func Shanghai() time.Time {
  return time.Now().In(TimeZoneSH)
}

func ShanghaiStr() string {
  return Shanghai().Format(DateTimeFormatSec)
}

func ShanghaiStrf(format string) string {
  return Shanghai().Format(format)
}

func UTC() time.Time {
  return time.Now().UTC()
}

func UTCStr() string {
  return UTC().Format(DateTimeFormatSec)
}

func UTCStrf(format string) string {
  return UTC().Format(format)
}

func DurationUntilTomorrow(t time.Time) time.Duration {
  t0 := t.Add(time.Hour * 24)
  t0 = time.Date(t0.Year(), t0.Month(), t0.Day(), 0, 0, 0, 0, t.Location())
  return t0.Sub(t)
}
