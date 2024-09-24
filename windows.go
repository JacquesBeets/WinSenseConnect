package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modWtsapi32              = windows.NewLazySystemDLL("Wtsapi32.dll")
	procWTSQueryUserToken    = modWtsapi32.NewProc("WTSQueryUserToken")
	procWTSEnumerateSessions = modWtsapi32.NewProc("WTSEnumerateSessionsW")
	procWTSFreeMemory        = modWtsapi32.NewProc("WTSFreeMemory")
)

type WTS_SESSION_INFO struct {
	SessionID     uint32
	WindowStation *uint16
	State         uint32
}

func getActiveSessionID() (uint32, error) {
	var count uint32
	var sessions uintptr
	ret, _, err := procWTSEnumerateSessions.Call(0, 0, 1, uintptr(unsafe.Pointer(&sessions)), uintptr(unsafe.Pointer(&count)))
	if ret == 0 {
		return 0, fmt.Errorf("WTSEnumerateSessions failed: %v", err)
	}
	defer procWTSFreeMemory.Call(sessions)

	for i := uint32(0); i < count; i++ {
		session := *(*WTS_SESSION_INFO)(unsafe.Pointer(sessions + uintptr(i)*unsafe.Sizeof(WTS_SESSION_INFO{})))
		if session.State == 0 { // WTSActive
			return session.SessionID, nil
		}
	}

	return 0, fmt.Errorf("no active session found")
}

func wtsQueryUserToken(session uint32, token *windows.Token) error {
	r1, _, e1 := procWTSQueryUserToken.Call(uintptr(session), uintptr(unsafe.Pointer(token)))
	if r1 == 0 {
		return error(e1)
	}
	return nil
}
