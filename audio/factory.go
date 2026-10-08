package audio

// NewCapture creates a platform-appropriate capture source.
func NewCapture(sampleRate, bufferSize int) CaptureSource {
	return newPlatformCapture(sampleRate, bufferSize)
}

// NewPlayback creates a platform-appropriate playback sink.
func NewPlayback(sampleRate, bufferSize int) PlaybackSink {
	return newPlatformPlayback(sampleRate, bufferSize)
}