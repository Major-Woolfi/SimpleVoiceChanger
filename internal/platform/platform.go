package internal/platform

type Platform int

const (
	PlatformWindows Platform = iota
	PlatformAndroid
	PlatformLinux
	PlatformMacOS
)

func Detect() Platform {
	return PlatformWindows
}
