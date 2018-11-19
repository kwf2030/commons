package crawler

import (
  "fmt"
  "html"
  "os"
  "runtime"
  "strings"
  "time"

  "github.com/kwf2030/commons/cdp"
  "github.com/kwf2030/commons/times"
  "github.com/rs/zerolog"
)

var (
  logFile  *os.File
  logger   *zerolog.Logger
  logLevel = zerolog.Disabled

  chrome *cdp.Chrome
)

/*func main() {
  crawler.SetLogLevel("debug")
  crawler.SetChrome("")
  e := crawler.GetRules().FromFiles([]string{"D:\\Workspace\\kwf2030\\commons\\crawler\\jd.yml"})
  if e != nil {
    panic(e)
  }
  ch := crawler.Enqueue(map[string]string{"1": "https://item.jd.com/11684158.html", "2": "https://item.jd.com/2165601.html"})
  data := <-ch
  fmt.Println(data)
}*/

func SetLogLevel(level string) {
  switch strings.ToLower(level) {
  case "debug":
    logLevel = zerolog.DebugLevel
  case "info":
    logLevel = zerolog.InfoLevel
  case "warn":
    logLevel = zerolog.WarnLevel
  case "error":
    logLevel = zerolog.ErrorLevel
  }
  initLogger()
}

func SetChrome(bin string, args ...string) {
  if bin == "" {
    switch runtime.GOOS {
    case "windows":
      bin = "C:/Program Files (x86)/Google/Chrome/Application/chrome.exe"
    case "linux":
      bin = "/usr/bin/google-chrome-stable"
    }
  }
  var e error
  chrome, e = cdp.Launch(bin, args...)
  if e != nil {
    panic(e)
  }
  tab, _ := chrome.NewTab()
  defer tab.Close()
  msg := tab.Call(cdp.Browser.GetVersion, nil)
  logger.Info().Msg(msg.Result["product"].(string))
}

func GetRules() *Rules {
  return allRules
}

func Stop() {
  if logFile != nil {
    logFile.Close()
  }
  if chrome != nil {
    tab, e := chrome.NewTab()
    if e == nil {
      tab.CallAsync(cdp.Browser.Close, nil)
    }
  }
}

func Enqueue(urls map[string]string) <-chan map[string]interface{} {
  ch := make(chan map[string]interface{})
  go func() {
    data := make(map[string]interface{}, len(urls))
    for id, addr := range urls {
      if addr == "" {
        continue
      }
      r := crawl(html.UnescapeString(addr))
      if len(r) > 0 {
        data[id] = r
      }
    }
    ch <- data
  }()
  return ch
}

func initLogger() {
  now := times.Now()
  if logger == nil {
    e := os.MkdirAll("log", os.ModePerm)
    if e != nil {
      panic(e)
    }
    next := now.Add(time.Hour * 24)
    next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
    time.AfterFunc(next.Sub(now), func() {
      logger.Info().Msg("create log file")
      go initLogger()
    })
  }
  zerolog.SetGlobalLevel(logLevel)
  zerolog.TimeFieldFormat = ""
  if logFile != nil {
    logFile.Close()
  }
  logFile, _ = os.Create(fmt.Sprintf("log/runner_%s.log", now.Format(times.DateFormat4)))
  lg := zerolog.New(logFile).Level(logLevel).With().Timestamp().Logger()
  logger = &lg
}
