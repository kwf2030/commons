package wechatbot

import (
  "bytes"
  "crypto/md5"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "mime"
  "mime/multipart"
  "net/http"
  "net/http/cookiejar"
  "net/url"
  "os"
  "path"
  "strconv"
  "strings"
  "time"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
  "golang.org/x/net/publicsuffix"
)

const (
  dateTimeFormat = "Mon Jan 02 2006 15:04:05 GMT-0700（中国标准时间）"

  chunkSize = 512 * 1024
)

var (
  verifyUrlPath        = "/webwxverifyuser"
  remarkUrlPath        = "/webwxoplog"
  signOutUrlPath       = "/webwxlogout"
  batchContactsUrlPath = "/webwxbatchgetcontact"
  sendTextUrlPath      = "/webwxsendmsg"
  sendEmotionUrlPath   = "/webwxsendemoticon"
  sendImageUrlPath     = "/webwxsendmsgimg"
  sendVideoUrlPath     = "/webwxsendvideomsg"
  uploadUrlPath        = "/webwxuploadmedia"
)

type req struct {
  bot    *Bot
  flow   *flow.Flow
  client *http.Client
  *session
}

func newReq() *req {
  jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
  sess := &session{}
  sess.reset()
  return &req{
    session: sess,
    flow:    flow.NewFlow(0),
    client: &http.Client{
      Jar:     jar,
      Timeout: time.Minute * 2,
    },
  }
}

func (r *req) initFlow() {
  qr := &qrReq{r}
  scan := &scanReq{r}
  redirect := &redirectReq{r}
  init := &initReq{r}
  notify := &notifyReq{r}
  contacts := &contactsReq{r}
  syn := &syncReq{r}
  r.flow.AddLast(qr, "qr")
  r.flow.AddLast(scan, "scan")
  r.flow.AddLast(redirect, "redirect")
  r.flow.AddLast(init, "init")
  r.flow.AddLast(notify, "notify")
  r.flow.AddLast(contacts, "contacts")
  r.flow.AddLast(syn, "sync")
}

func (r *req) cookie(key string) string {
  if key == "" {
    return ""
  }
  addr, _ := url.Parse(r.BaseUrl)
  arr := r.client.Jar.Cookies(addr)
  for _, c := range arr {
    if c.Name == key {
      return c.Value
    }
  }
  return ""
}

func (r *req) DownloadQRCode(dst string) (string, error) {
  resp, e := http.Get(r.QRCodeUrl)
  if e != nil {
    return "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return "", ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return "", e
  }
  dumpToFile("DownloadQRCode_"+times.NowStrf(times.DateTimeMsFormat5), body)
  if dst == "" {
    dst = path.Join(os.TempDir(), "wechatbot_qrcode.jpg")
  }
  e = ioutil.WriteFile(dst, body, os.ModePerm)
  if e != nil {
    return "", e
  }
  return dst, nil
}

func (r *req) DownloadAvatar(dst string) (string, error) {
  resp, e := r.client.Get(r.AvatarUrl)
  if e != nil {
    return "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return "", ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return "", e
  }
  dumpToFile("DownloadAvatar_"+times.NowStrf(times.DateTimeMsFormat5), body)
  if dst == "" {
    dst = path.Join(os.TempDir(), fmt.Sprintf("wechatbot_%d.jpg", r.Uin))
  }
  e = ioutil.WriteFile(dst, body, os.ModePerm)
  if e != nil {
    return "", e
  }
  return dst, nil
}

func (r *req) Verify(toUserName, ticket string) ([]byte, error) {
  addr, _ := url.Parse(r.BaseUrl + verifyUrlPath)
  q := addr.Query()
  q.Set("r", timestampString13())
  q.Set("pass_ticket", r.PassTicket)
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 8)
  m["BaseRequest"] = r.BaseReq
  m["skey"] = r.SKey
  m["Opcode"] = 3
  m["SceneListCount"] = 1
  m["SceneList"] = []int{33}
  m["VerifyContent"] = ""
  m["VerifyUserListSize"] = 1
  m["VerifyUserList"] = []map[string]string{
    {
      "Value":            toUserName,
      "VerifyUserTicket": ticket,
    },
  }
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.Referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("Verify_"+times.NowStrf(times.DateTimeMsFormat5), body)
  return body, nil
}

func (r *req) Remark(toUserName, remark string) ([]byte, error) {
  addr, _ := url.Parse(r.BaseUrl + remarkUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.PassTicket)
  addr.RawQuery = q.Encode()
  m := make(map[string]interface{}, 4)
  m["BaseRequest"] = r.BaseReq
  m["UserName"] = toUserName
  m["CmdId"] = 2
  m["RemarkName"] = remark
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.Referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("Remark_"+times.NowStrf(times.DateTimeMsFormat5), body)
  return body, nil
}

func (r *req) GetContacts(toUserNames ...string) ([]byte, error) {
  addr, _ := url.Parse(r.BaseUrl + batchContactsUrlPath)
  q := addr.Query()
  q.Set("type", "ex")
  q.Set("r", timestampString13())
  addr.RawQuery = q.Encode()
  arr := make([]map[string]string, 0, len(toUserNames))
  for _, userName := range toUserNames {
    m := make(map[string]string, 2)
    m["UserName"] = userName
    m["EncryChatRoomId"] = ""
    arr = append(arr, m)
  }
  m := make(map[string]interface{}, 3)
  m["BaseRequest"] = r.BaseReq
  m["Count"] = len(toUserNames)
  m["List"] = arr
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.Referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("GetContacts_"+times.NowStrf(times.DateTimeMsFormat5), body)
  return body, nil
}

func (r *req) SignOut() ([]byte, error) {
  addr, _ := url.Parse(r.BaseUrl + signOutUrlPath)
  q := addr.Query()
  q.Set("redirect", "1")
  q.Set("type", "1")
  q.Set("skey", r.SKey)
  addr.RawQuery = q.Encode()
  form := url.Values{}
  form.Set("sid", r.Sid)
  form.Set("uin", strconv.FormatInt(r.Uin, 10))
  req, _ := http.NewRequest("POST", addr.String(), strings.NewReader(form.Encode()))
  req.Header.Set("Referer", r.Referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("SignOut_"+times.NowStrf(times.DateTimeMsFormat5), body)
  return body, nil
}

func (r *req) SendText(toUserName, text string) ([]byte, error) {
  addr, _ := url.Parse(r.BaseUrl + sendTextUrlPath)
  q := addr.Query()
  q.Set("pass_ticket", r.PassTicket)
  addr.RawQuery = q.Encode()
  n, _ := strconv.ParseInt(timestampString13(), 10, 32)
  s := strconv.FormatInt(n<<4, 10) + timestampStringR(4)
  params := map[string]interface{}{
    "Type":         MsgText,
    "Content":      text,
    "FromUserName": r.UserName,
    "ToUserName":   toUserName,
    "LocalID":      s,
    "ClientMsgId":  s,
  }
  m := make(map[string]interface{}, 3)
  m["BaseRequest"] = r.BaseReq
  m["Scene"] = 0
  m["Msg"] = params
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.Referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("SendText_"+times.NowStrf(times.DateTimeMsFormat5), body)
  return body, nil
}

func (r *req) SendMedia(toUserName, mediaId string, msgType int, sendUrlPath string) ([]byte, error) {
  addr, _ := url.Parse(r.BaseUrl + sendUrlPath)
  q := addr.Query()
  q.Set("fun", "async")
  q.Set("f", "json")
  q.Set("pass_ticket", r.PassTicket)
  addr.RawQuery = q.Encode()
  n, _ := strconv.ParseInt(timestampString13(), 10, 32)
  s := strconv.FormatInt(n<<4, 10) + timestampStringR(4)
  params := map[string]interface{}{
    "Type":         msgType,
    "MediaId":      mediaId,
    "FromUserName": r.UserName,
    "ToUserName":   toUserName,
    "LocalID":      s,
    "ClientMsgId":  s,
    "Content":      "",
  }
  m := make(map[string]interface{}, 3)
  m["BaseRequest"] = r.BaseReq
  m["Scene"] = 0
  m["Msg"] = params
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.Referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  dumpToFile("SendMedia_"+times.NowStrf(times.DateTimeMsFormat5), body)
  return body, nil
}

// data是上传的数据，如果大于chunk则按chunk分块上传，
// filename是文件名（非文件路径，用来检测文件类型和设置上传文件名，如1.png）
func (r *req) UploadMedia(toUserName string, data []byte, filename string) (string, error) {
  l := len(data)
  addr, _ := url.Parse(r.BaseUrl + uploadUrlPath)
  addr.Host = "file." + addr.Host
  q := addr.Query()
  q.Set("f", "json")
  addr.RawQuery = q.Encode()

  mimeType := "application/octet-stream"
  i := strings.LastIndex(filename, ".")
  if i != -1 {
    mt := mime.TypeByExtension(filename[i:])
    if mt != "" {
      mimeType = mt
    }
  }

  mediaType := "doc"
  switch mimeType[:strings.Index(mimeType, "/")] {
  case "image":
    mediaType = "pic"
  case "video":
    mediaType = "video"
  }

  hash := fmt.Sprintf("%x", md5.Sum(data))
  n, _ := strconv.ParseInt(timestampString13(), 10, 32)
  s := strconv.FormatInt(n<<4, 10) + timestampStringR(4)
  m := make(map[string]interface{}, 10)
  m["BaseRequest"] = r.BaseReq
  m["UploadType"] = 2
  m["ClientMediaId"] = s
  m["TotalLen"] = l
  m["DataLen"] = l
  m["StartPos"] = 0
  m["MediaType"] = 4
  m["FromUserName"] = r.UserName
  m["ToUserName"] = toUserName
  m["FileMd5"] = hash
  payload, _ := json.Marshal(m)

  info := &uploadInfo{
    addr:         addr.String(),
    filename:     filename,
    md5:          hash,
    mimeType:     mimeType,
    mediaType:    mediaType,
    payload:      string(payload),
    fromUserName: r.UserName,
    toUserName:   toUserName,
    dataTicket:   r.cookie("webwx_data_ticket"),
    totalLen:     l,
    wuFile:       r.WuFile,
    chunks:       0,
    chunk:        0,
    data:         nil,
  }
  defer func() { r.WuFile++ }()

  var mediaId string
  var err error
  if l <= chunkSize {
    info.data = data
    mediaId, err = r.uploadChunk(info)
  } else {
    m := l / chunkSize
    n := l % chunkSize
    if n == 0 {
      info.chunks = m
    } else {
      info.chunks = m + 1
    }
    for i := 0; i < m; i++ {
      s := i * chunkSize
      e := s + chunkSize
      info.chunk = i
      info.data = data[s:e]
      mediaId, err = r.uploadChunk(info)
      if err != nil {
        break
      }
    }
    if n != 0 && err == nil {
      info.chunk++
      info.data = data[l-n:]
      mediaId, err = r.uploadChunk(info)
    }
  }
  return mediaId, err
}

func (r *req) uploadChunk(info *uploadInfo) (string, error) {
  var buf bytes.Buffer
  w := multipart.NewWriter(&buf)
  defer w.Close()
  w.WriteField("id", fmt.Sprintf("WU_FILE_%d", info.wuFile))
  w.WriteField("name", info.filename)
  w.WriteField("type", info.mimeType)
  w.WriteField("lastModifiedDate", times.Now().Add(time.Hour * -24).Format(dateTimeFormat))
  w.WriteField("size", strconv.Itoa(info.totalLen))
  if info.chunks > 0 {
    w.WriteField("chunks", strconv.Itoa(info.chunks))
    w.WriteField("chunk", strconv.Itoa(info.chunk))
  }
  w.WriteField("mediatype", info.mediaType)
  w.WriteField("uploadmediarequest", info.payload)
  w.WriteField("webwx_data_ticket", info.dataTicket)
  w.WriteField("pass_ticket", r.PassTicket)
  fw, e := w.CreateFormFile("filename", info.filename)
  if e != nil {
    return "", e
  }
  if _, e = fw.Write(info.data); e != nil {
    return "", e
  }

  req, _ := http.NewRequest("POST", info.addr, &buf)
  req.Header.Set("Referer", r.Referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", w.FormDataContentType())
  resp, e := r.client.Do(req)
  if e != nil {
    return "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return "", ErrReq
  }
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return "", e
  }
  dumpToFile("uploadChunk_"+times.NowStrf(times.DateTimeMsFormat5), body)
  mediaId, _ := jsonparser.GetString(body, "MediaId")
  return mediaId, nil
}

type uploadInfo struct {
  addr         string
  filename     string
  md5          string
  mimeType     string
  mediaType    string
  payload      string
  fromUserName string
  toUserName   string
  dataTicket   string
  totalLen     int
  wuFile       int
  chunks       int
  chunk        int
  data         []byte
}

type session struct {
  SyncCheckHost string
  Host          string
  Referer       string
  BaseUrl       string
  State         int
  UUID          string
  QRCodeUrl     string
  RedirectUrl   string
  SKey          string
  Sid           string
  Uin           int64
  PassTicket    string
  BaseReq       *baseReq
  UserName      string
  AvatarUrl     string
  SyncKeys      *syncKeys
  WuFile        int
}

func (s *session) reset() {
  s.SyncCheckHost = "webpush.weixin.qq.com"
  s.Host = "wx.qq.com"
  s.Referer = "https://wx.qq.com/"
  s.BaseUrl = "https://wx.qq.com/cgi-bin/mmwebwx-bin"
  s.State = StateCreated
  s.UUID = ""
  s.QRCodeUrl = ""
  s.RedirectUrl = ""
  s.BaseReq = nil
  s.SKey = ""
  s.Sid = ""
  s.Uin = 0
  s.PassTicket = ""
  s.UserName = ""
  s.AvatarUrl = ""
  s.SyncKeys = nil
  s.WuFile = 0
}
