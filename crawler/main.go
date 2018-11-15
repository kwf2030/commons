package crawler

import (
  "fmt"
  "io/ioutil"
  "os"
  "runtime"
  "time"

  "github.com/kwf2030/commons/cdp"
  "github.com/kwf2030/commons/times"
  "github.com/rs/zerolog"
  "gopkg.in/yaml.v2"
)

var (
  logFile  *os.File
  logger   *zerolog.Logger
  logLevel = zerolog.Disabled

  chrome *cdp.Chrome
)

func SetLogLevel(level string) {
  switch level {
  case "debug":
    logLevel = zerolog.DebugLevel
  case "info":
    logLevel = zerolog.InfoLevel
  case "warn":
    logLevel = zerolog.WarnLevel
  case "error":
    logLevel = zerolog.ErrorLevel
  }
}

func Start() {
  initLogger()
  initChrome()
}

func Stop() {
  logFile.Close()
  tab, e := chrome.NewTab()
  if e == nil {
    tab.CallAsync(cdp.Browser.Close, nil)
  }
}

func Enqueue(urls []string) {

}

func initLogger() {
  e := os.MkdirAll("log", os.ModePerm)
  if e != nil {
    panic(e)
  }
  zerolog.SetGlobalLevel(zerolog.InfoLevel)
  zerolog.TimeFieldFormat = ""
  if logFile != nil {
    logFile.Close()
  }
  now := times.Now()
  logFile, _ = os.Create(fmt.Sprintf("log/runner_%s.log", now.Format(times.DateFormat4)))
  lg := zerolog.New(logFile).Level(zerolog.InfoLevel).With().Timestamp().Logger()
  logger = &lg
  next := now.Add(time.Hour * 24)
  next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
  time.AfterFunc(next.Sub(now), func() {
    logger.Info().Msg("create log file")
    go initLogger()
  })
}

func initChrome() {
  var e error
  c := &struct {
    Exec string
    Args []string
  }{}
  f := "chrome.yml"
  _, e = os.Stat(f)
  if e == nil {
    data, err := ioutil.ReadFile(f)
    if err == nil {
      yaml.Unmarshal(data, c)
    }
  }
  if c.Exec == "" {
    switch runtime.GOOS {
    case "windows":
      c.Exec = "C:/Program Files (x86)/Google/Chrome/Application/chrome.exe"
    case "linux":
      c.Exec = "/usr/bin/google-chrome-stable"
    }
  }
  chrome, e = cdp.Launch(c.Exec, c.Args...)
  if e != nil {
    panic(e)
  }
  tab, _ := chrome.NewTab()
  defer tab.Close()
  msg := tab.Call(cdp.Browser.GetVersion, nil)
  logger.Info().Msg(msg.Result["product"].(string))
}
