//go:build !linux

package dns

// EnsureWSLResolver has nothing to do off Linux, where there is no WSL.
func EnsureWSLResolver() error { return nil }

// WSLResolverMissing is false off Linux, where there is no WSL.
func WSLResolverMissing() bool { return false }
