//go:build !windows

package main

import "os"

// The Windows-only helpers; the package builds elsewhere so it is vetted and
// tested with the rest of lerd.
func redirectStderr(*os.File) {}

func claimAgent() bool { return true }
