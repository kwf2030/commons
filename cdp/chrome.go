package cdp

import (
  "encoding/json"
  "errors"
  "fmt"
  "net/http"
  "os"
  "os/exec"
  "strings"
  "sync"
  "time"
)

type Chrome struct {
  // ChromeDevToolsProtocol的API地址（http://host:port/json）
  Endpoint string
  Process  *os.Process
}

func Launch(bin string, args ...string) (*Chrome, error) {
  if bin == "" {
    return nil, errors.New("empty <bin>")
  }
  _, e := exec.LookPath(bin)
  if e != nil {
    return nil, e
  }
  var port string
  for _, v := range args {
    if strings.Contains(v, "--remote-debugging-port") {
      arr := strings.Split(v, "=")
      if len(arr) != 2 {
        return nil, errors.New("invalid '--remote-debugging-port'")
      }
      port = strings.TrimSpace(arr[1])
      break
    }
  }
  if port == "" {
    port = "9222"
    args = append(args, fmt.Sprintf("--remote-debugging-port=%s", port))
  }
  cmd := exec.Command(bin, args...)
  e = cmd.Start()
  if e != nil {
    return nil, e
  }
  c := &Chrome{fmt.Sprintf("http://127.0.0.1:%s/json", port), cmd.Process}
  if ok := c.waitForStarted(time.Second * 10); !ok {
    return nil, errors.New("chrome not started")
  }
  return c, nil
}

func Connect(host string, port int) (*Chrome, error) {
  if host == "" || port <= 0 {
    return nil, errors.New("invalid <host>/<port>")
  }
  return &Chrome{fmt.Sprintf("http://%s:%d/json", host, port), nil}, nil
}

func (c *Chrome) NewTab() (*Tab, error) {
  meta := &tabMeta{}
  resp, e := http.Get(c.Endpoint + "/new")
  if e != nil {
    return nil, e
  }
  e = json.NewDecoder(resp.Body).Decode(meta)
  resp.Body.Close()
  if e != nil {
    return nil, e
  }
  if meta.Id == "" || meta.WebSocketDebuggerUrl == "" {
    return nil, errors.New("NewTab receives empty Id/WebSocketDebuggerUrl")
  }
  t := &Tab{
    chrome:            c,
    meta:              meta,
    closeChan:         make(chan struct{}),
    sendChan:          make(chan *Message, 1),
    C:                 make(chan *Message, 2),
    eventsAndMessages: sync.Map{},
  }
  t.conn, e = t.wsConnect()
  if e != nil {
    t.Close()
    return nil, e
  }
  go t.wsRead()
  go t.wsWrite()
  return t, nil
}

func (c *Chrome) waitForStarted(timeout time.Duration) bool {
  client := &http.Client{Timeout: time.Second}
  t := time.After(timeout)
  for {
    select {
    case <-t:
      return false
    default:
      resp, e := client.Get(c.Endpoint)
      if e != nil {
        break
      }
      drain(resp.Body)
      return true
    }
  }
}
