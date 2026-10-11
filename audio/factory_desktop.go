//go:build !android

package audio

func newPlatformCapture(sampleRate, bufferSize int, deviceName string) CaptureSource {
	c, err := NewPortAudioCapture(sampleRate, bufferSize, deviceName)
	if err != nil {
		return NewDefaultCapture()
	}
	return c
}

func newPlatformPlayback(sampleRate, bufferSize int, deviceName string) PlaybackSink {
	p, err := NewPortAudioPlayback(sampleRate, bufferSize, deviceName)
	if err != nil {
		return NewDefaultPlayback()
	}
	return p
}
