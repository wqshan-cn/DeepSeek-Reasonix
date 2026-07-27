//go:build windows

package main

import (
	"unsafe"

	syswindows "golang.org/x/sys/windows"
)

const (
	desktopPetTransparencyR uint8 = 1
	desktopPetTransparencyG uint8 = 2
	desktopPetTransparencyB uint8 = 3

	desktopPetGWLExStyle  = ^uintptr(19)
	desktopPetWSExLayered = uintptr(0x00080000)
	desktopPetLWAColorKey = uintptr(0x00000001)
)

var (
	desktopPetUser32                     = syswindows.NewLazySystemDLL("user32.dll")
	desktopPetFindWindowW                = desktopPetUser32.NewProc("FindWindowW")
	desktopPetGetWindowLongPtrW          = desktopPetUser32.NewProc("GetWindowLongPtrW")
	desktopPetSetWindowLongPtrW          = desktopPetUser32.NewProc("SetWindowLongPtrW")
	desktopPetSetLayeredWindowAttributes = desktopPetUser32.NewProc("SetLayeredWindowAttributes")
)

func enableDesktopPetTransparency() {
	className, err := syswindows.UTF16PtrFromString("ReasonixDesktopPet")
	if err != nil {
		return
	}
	hwnd, _, _ := desktopPetFindWindowW.Call(uintptr(unsafe.Pointer(className)), 0)
	if hwnd == 0 {
		return
	}

	exStyle, _, _ := desktopPetGetWindowLongPtrW.Call(hwnd, desktopPetGWLExStyle)
	_, _, _ = desktopPetSetWindowLongPtrW.Call(
		hwnd,
		desktopPetGWLExStyle,
		exStyle|desktopPetWSExLayered,
	)

	colourKey := uintptr(desktopPetTransparencyR) |
		uintptr(desktopPetTransparencyG)<<8 |
		uintptr(desktopPetTransparencyB)<<16
	_, _, _ = desktopPetSetLayeredWindowAttributes.Call(
		hwnd,
		colourKey,
		255,
		desktopPetLWAColorKey,
	)
}
