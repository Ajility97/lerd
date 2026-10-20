package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var procWNetGetConnection = windows.NewLazySystemDLL("mpr.dll").NewProc("WNetGetConnectionW")

// driveShare asks Windows which share drive ("Z:") is mapped to.
func driveShare(drive string) string {
	buf := make([]uint16, windows.MAX_LONG_PATH)
	n := uint32(len(buf))
	r, _, _ := procWNetGetConnection.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(drive))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&n)),
	)
	if r != 0 {
		return ""
	}
	return windows.UTF16ToString(buf)
}
