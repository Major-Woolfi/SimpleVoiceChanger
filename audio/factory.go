package audio

// NewCapture creates a platform-appropriate capture source. An empty
// deviceName selects the system default input device.
func NewCapture(sampleRate, bufferSize int, deviceName string) CaptureSource {
	return newPlatformCapture(sampleRate, bufferSize, deviceName)
}

// NewPlayback creates a platform-appropriate playback sink. An empty
// deviceName selects the system default output device.
func NewPlayback(sampleRate, bufferSize int, deviceName string) PlaybackSink {
	return newPlatformPlayback(sampleRate, bufferSize, deviceName)
}
