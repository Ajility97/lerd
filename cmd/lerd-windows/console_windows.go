package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// redirectStderr sends this process's stderr, Go's and the Win32 handle, to f.
func redirectStderr(f *os.File) {
	os.Stderr = f
	windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(f.Fd())) //nolint:errcheck
}

// claimAgent reports false when another agent in this session holds the mutex,
// so a second one, started again by the desktop app, steps aside from port 53.
// An elevated agent's mutex cannot be opened unelevated, which also means taken.
func claimAgent() bool {
	return claimMutex(`Local\lerd-agent`)
}

func claimMutex(name string) bool {
	_, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(name))
	return err != windows.ERROR_ALREADY_EXISTS && err != windows.ERROR_ACCESS_DENIED
}
