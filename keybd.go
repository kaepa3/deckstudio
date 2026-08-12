package main

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procSendInput    = user32.NewProc("SendInput")
	procMapVirtualKey = user32.NewProc("MapVirtualKeyW")
)

const (
	INPUT_KEYBOARD = 1

	KEYEVENTF_EXTENDEDKEY = 0x0001
	KEYEVENTF_KEYUP       = 0x0002
	KEYEVENTF_UNICODE     = 0x0004
	KEYEVENTF_SCANCODE    = 0x0008

	MAPVK_VK_TO_VSC = 0
)

type KEYBDINPUT struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
	_pad [8]byte // 64-bit alignment padding for Win32 SendInput struct
}

var vkMap = map[string]uint16{
	// Control / Modifiers
	"ctrl":    0x11, // VK_CONTROL
	"control": 0x11,
	"shift":   0x10, // VK_SHIFT
	"alt":     0x12, // VK_MENU
	"win":     0x5B, // VK_LWIN
	"cmd":     0x5B,

	// Basic keys
	"tab":       0x09, // VK_TAB
	"enter":     0x0D, // VK_RETURN
	"return":    0x0D,
	"escape":    0x1B, // VK_ESCAPE
	"esc":       0x1B,
	"space":     0x20, // VK_SPACE
	"backspace": 0x08, // VK_BACK
	"delete":    0x46, // VK_DELETE

	// Arrows
	"up":    0x26,
	"down":  0x28,
	"left":  0x25,
	"right": 0x27,

	// Media Keys
	"volume_up":       0xAF, // VK_VOLUME_UP
	"volume_down":     0xAE, // VK_VOLUME_DOWN
	"volume_mute":     0xAD, // VK_VOLUME_MUTE
	"media_next":      0xB0, // VK_MEDIA_NEXT_TRACK
	"media_prev":      0xB1, // VK_MEDIA_PREV_TRACK
	"media_play_pause": 0xB3, // VK_MEDIA_PLAY_PAUSE

	// Symbols
	"/":  0xBF, // VK_OEM_2
	"`":  0xC0, // VK_OEM_3
	"-":  0xBD, // VK_OEM_MINUS
	"=":  0xBB, // VK_OEM_PLUS
	"[":  0xDB, // VK_OEM_4
	"]":  0xDD, // VK_OEM_6
	"\\": 0xDC, // VK_OEM_5
	";":  0xBA, // VK_OEM_1
	"'":  0xDE, // VK_OEM_7
	",":  0xBC, // VK_OEM_COMMA
	".":  0xBE, // VK_OEM_PERIOD
}

func getVKCode(k string) uint16 {
	kLower := strings.ToLower(strings.TrimSpace(k))
	if vk, ok := vkMap[kLower]; ok {
		return vk
	}

	// Single char A-Z, 0-9
	if len(kLower) == 1 {
		ch := kLower[0]
		if ch >= 'a' && ch <= 'z' {
			return uint16('A' + (ch - 'a'))
		}
		if ch >= '0' && ch <= '9' {
			return uint16(ch)
		}
	}

	// F1 - F12
	if strings.HasPrefix(kLower, "f") && len(kLower) <= 3 {
		var fNum int
		fmt.Sscanf(kLower, "f%d", &fNum)
		if fNum >= 1 && fNum <= 12 {
			return uint16(0x70 + (fNum - 1))
		}
	}

	return 0
}

func sendInputKeys(vks []uint16, down bool) {
	if len(vks) == 0 {
		return
	}

	var inputs []INPUT
	for _, vk := range vks {
		var flags uint32 = 0
		if !down {
			flags |= KEYEVENTF_KEYUP
		}

		// Extended keys (media keys, arrows, win, etc.)
		if vk >= 0x25 && vk <= 0x28 || vk >= 0xAD && vk <= 0xB3 || vk == 0x5B {
			flags |= KEYEVENTF_EXTENDEDKEY
		}

		inp := INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				Vk:    vk,
				Flags: flags,
			},
		}
		inputs = append(inputs, inp)
	}

	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
}

func ExecuteButtonAction(btn *ButtonConfig) error {
	if btn == nil {
		return fmt.Errorf("button is nil")
	}

	var vks []uint16
	for _, keyStr := range btn.Keys {
		vk := getVKCode(keyStr)
		if vk != 0 {
			vks = append(vks, vk)
		}
	}

	if len(vks) == 0 {
		return fmt.Errorf("no valid keys for button: %s", btn.ID)
	}

	// Press keys in order
	for _, vk := range vks {
		sendInputKeys([]uint16{vk}, true)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(30 * time.Millisecond)

	// Release keys in reverse order
	for i := len(vks) - 1; i >= 0; i-- {
		sendInputKeys([]uint16{vks[i]}, false)
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}
