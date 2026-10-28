package audio

import "strings"

// builtinLoopbackMarkers: names of Windows recording devices that capture the
// system playback mix. This is the built-in "second microphone" — a component
// of the audio driver (Realtek, Conexant, etc.), not a third-party program.
//
// When the voice changer writes processed audio to the default output (speakers)
// the user picks the matching capture device (e.g. "Stereo Mix") in Discord or
// any chat app, and the processed voice becomes their microphone.
//
// Common names across languages and vendors:
var builtinLoopbackMarkers = []string{
	"stereo mix",
	"стерео микш",
	"what u hear",
	"what you hear",
	"what-u-hear",
	"mix out",
	"summing",
	"udwa",
	"aux mix",
	"wave out",
	"loopback",
}

// FindBuiltinLoopback scans the available capture (input) devices and returns
// the first one that looks like a system mix recorder (Stereo Mix, What U
// Hear, etc.). It returns an empty string when no such device is present.
func FindBuiltinLoopback() string {
	return findBuiltinIn(ListInputDevices())
}

func findBuiltinIn(devices []string) string {
	for _, name := range devices {
		lower := strings.ToLower(name)
		for _, marker := range builtinLoopbackMarkers {
			if strings.Contains(lower, marker) {
				return name
			}
		}
	}
	return ""
}

// IsBuiltinLoopback reports whether the named capture device is a system mix
// recorder (Stereo Mix, What U Hear, etc.).
func IsBuiltinLoopback(name string) bool {
	lower := strings.ToLower(name)
	for _, marker := range builtinLoopbackMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
