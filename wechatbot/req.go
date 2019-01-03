package wechatbot

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/kwf2030/commons/times"
  "io/ioutil"
  "net/http"
  "net/url"
  "os"
  "path"
  "strconv"
  "strings"
)

var (
  verifyUrlPath   = "/webwxverifyuser"
  remarkUrlPath   = "/webwxoplog"
  signOutUrlPath  = "/webwxlogout"
  contactsUrlPath = "/webwxbatchgetcontact"
)

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

func (r *req) GetContacts(userNames ...string) ([]byte, error) {
  addr, _ := url.Parse(r.BaseUrl + contactsUrlPath)
  q := addr.Query()
  q.Set("type", "ex")
  q.Set("r", timestampString13())
  addr.RawQuery = q.Encode()
  arr := make([]map[string]string, 0, len(userNames))
  for _, userName := range userNames {
    m := make(map[string]string, 2)
    m["UserName"] = userName
    m["EncryChatRoomId"] = ""
    arr = append(arr, m)
  }
  m := make(map[string]interface{}, 3)
  m["BaseRequest"] = r.BaseReq
  m["Count"] = len(userNames)
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
