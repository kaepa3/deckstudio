package main

import (
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")

	kernel32                      = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procCloseHandle               = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
)

type ActiveAppInfo struct {
	AppName string `json:"app_name"`
	Title   string `json:"title"`
}

func GetActiveAppInfo() (ActiveAppInfo, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ActiveAppInfo{}, nil
	}

	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	if pid == 0 {
		return ActiveAppInfo{}, nil
	}

	// Window Title
	var titleBuf [512]uint16
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), uintptr(len(titleBuf)))
	title := syscall.UTF16ToString(titleBuf[:])

	// Process Handle
	hProcess, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_LIMITED_INFORMATION),
		uintptr(0),
		uintptr(pid),
	)
	if hProcess == 0 {
		return ActiveAppInfo{AppName: "", Title: title}, nil
	}
	defer procCloseHandle.Call(hProcess)

	var exePathBuf [1024]uint16
	var size uint32 = uint32(len(exePathBuf))

	ret, _, _ := procQueryFullProcessImageName.Call(
		hProcess,
		uintptr(0),
		uintptr(unsafe.Pointer(&exePathBuf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret == 0 {
		return ActiveAppInfo{AppName: "", Title: title}, nil
	}

	exePath := syscall.UTF16ToString(exePathBuf[:size])
	appName := filepath.Base(exePath)

	return ActiveAppInfo{
		AppName: strings.ToLower(appName),
		Title:   title,
	}, nil
}
