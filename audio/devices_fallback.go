//go:build !portaudio

package audio

// ListInputDevices returns no devices when the PortAudio backend is not built.
func ListInputDevices() []string { return nil }

// ListOutputDevices returns no devices when the PortAudio backend is not built.
func ListOutputDevices() []string { return nil }
