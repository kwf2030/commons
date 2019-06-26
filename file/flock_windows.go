package file

import (
  "os"
  "syscall"
  "time"
  "unsafe"

  "github.com/kwf2030/commons/base"
)

var errLockViolation syscall.Errno = 0x21

var (
  modkernel32      = syscall.NewLazyDLL("kernel32.dll")
  procLockFileEx   = modkernel32.NewProc("LockFileEx")
  procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

func lockFileEx(h syscall.Handle, flags, reserved, locklow, lockhigh uint32, ol *syscall.Overlapped) error {
  r, _, e := procLockFileEx.Call(uintptr(h), uintptr(flags), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)))
  if r == 0 {
    return e
  }
  return nil
}

func unlockFileEx(h syscall.Handle, reserved, locklow, lockhigh uint32, ol *syscall.Overlapped) error {
  r, _, e := procUnlockFileEx.Call(uintptr(h), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)), 0)
  if r == 0 {
    return e
  }
  return nil
}

// 文件锁（同时只有一个进程能持有锁），锁定期间再次调用Lock函数会在超时后返回timeout，
// 注意其他进程是否可读写是文件打开方式决定的，与锁无关，
// 例如使用os.O_WRONLY打开文件其他进程就无法写，使用os.O_RDONLY其他进程就可写
func Lock(f *os.File, timeout time.Duration) error {
  if f == nil || timeout <= 0 {
    return base.ErrInvalidArgument
  }
  t := time.Now()
  for {
    var m uint32 = (1 << 32) - 1
    e := lockFileEx(syscall.Handle(f.Fd()), 3, 0, 1, 0, &syscall.Overlapped{
      Offset:     m,
      OffsetHigh: m,
    })
    if e == nil {
      return nil
    }
    if e != errLockViolation {
      return e
    }
    if time.Since(t) > timeout {
      return base.ErrTimeout
    }
    time.Sleep(time.Millisecond * 100)
  }
}

func Unlock(f *os.File) error {
  if f == nil {
    return base.ErrInvalidArgument
  }
  var m uint32 = (1 << 32) - 1
  return unlockFileEx(syscall.Handle(f.Fd()), 0, 1, 0, &syscall.Overlapped{
    Offset:     m,
    OffsetHigh: m,
  })
}
