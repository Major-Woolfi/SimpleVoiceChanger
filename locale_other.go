//go:build !windows

package main

import (
	"os"
	"strings"
)

// detectSystemLocale inspects the POSIX locale environment variables and
// maps the language part to a supported locale, or an empty string.
func detectSystemLocale() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		value := os.Getenv(key)
		if value == "" {
			continue
		}
		code := strings.ToLower(value)
		code = strings.SplitN(code, ".", 2)[0]
		code = strings.SplitN(code, "_", 2)[0]
		code = strings.TrimSpace(code)
		switch code {
		case "ru", "uk", "de", "fr", "es", "zh":
			return code
		}
	}
	return "en"
}
