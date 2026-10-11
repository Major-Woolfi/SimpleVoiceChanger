//go:build windows

package main

import "syscall"

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	getUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
)

// detectSystemLocale returns the two-letter UI locale code of the current
// Windows user, mapped to a supported locale, or an empty string.
func detectSystemLocale() string {
	ret, _, _ := getUserDefaultUILanguage.Call()
	primary := uint16(ret) & 0x3ff
	switch primary {
	case 0x19:
		return "ru"
	case 0x22:
		return "uk"
	case 0x07:
		return "de"
	case 0x0c:
		return "fr"
	case 0x0a:
		return "es"
	case 0x04:
		return "zh"
	default:
		return "en"
	}
}
