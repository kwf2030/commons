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
  "net/url"
  "strconv"
  "strings"
  "time"

  "github.com/buger/jsonparser"
  "github.com/kwf2030/commons/times"
)

const (
  sendTextUrlPath    = "/webwxsendmsg"
  sendEmotionUrlPath = "/webwxsendemoticon"
  sendImageUrlPath   = "/webwxsendmsgimg"
  sendVideoUrlPath   = "/webwxsendvideomsg"
  uploadUrlPath      = "/webwxuploadmedia"
)

const dateTimeFormat = "Mon Jan 02 2006 15:04:05 GMT-0700（中国标准时间）"

const chunk = 512 * 1024

var jsonPathMediaId = "MediaId"

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
  if l <= chunk {
    info.data = data
    mediaId, err = r.uploadChunk(info)
  } else {
    m := l / chunk
    n := l % chunk
    if n == 0 {
      info.chunks = m
    } else {
      info.chunks = m + 1
    }
    for i := 0; i < m; i++ {
      s := i * chunk
      e := s + chunk
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
  mediaId, _ := jsonparser.GetString(body, jsonPathMediaId)
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
