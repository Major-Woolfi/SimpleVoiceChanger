package platform

import (
	"runtime"
)

// Platform enumerates supported operating systems.
type Platform int

const (
	PlatformWindows Platform = iota
	PlatformAndroid
	PlatformLinux
	PlatformMacOS
)

// Detect returns the platform for the current build.
func Detect() Platform {
	switch runtime.GOOS {
	case "android":
		return PlatformAndroid
	case "linux":
		return PlatformLinux
	case "darwin":
		return PlatformMacOS
	default:
		return PlatformWindows
	}
}

// String returns a human-readable platform name.
func (p Platform) String() string {
	switch p {
	case PlatformWindows:
		return "windows"
	case PlatformAndroid:
		return "android"
	case PlatformLinux:
		return "linux"
	case PlatformMacOS:
		return "macos"
	default:
		return "unknown"
	}
}

// IsGUI returns whether the platform supports a graphical user interface.
func (p Platform) IsGUI() bool {
	return p != PlatformAndroid
}
