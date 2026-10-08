package core

import "os"

// DefaultDirectories returns the standard directory layout for the app.
// Keys: "config" (app config dir), "presets" (user presets dir).
func DefaultDirectories() map[string]string {
	base := appDataDir()
	return map[string]string{
		"config":  base,
		"presets": base + "/presets",
	}
}

func appDataDir() string {
	if d, err := os.UserConfigDir(); err == nil {
		return d + "/SimpleVoiceChanger"
	}
	return "./SimpleVoiceChanger"
}