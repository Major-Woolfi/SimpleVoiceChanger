//go:build android

package audio

func newPlatformCapture(sampleRate, bufferSize int) CaptureSource {
	c, err := NewAndroidCapture(sampleRate, bufferSize)
	if err != nil {
		return NewDefaultCapture()
	}
	return c
}

func newPlatformPlayback(sampleRate, bufferSize int) PlaybackSink {
	p, err := NewAndroidPlayback(sampleRate, bufferSize)
	if err != nil {
		return NewDefaultPlayback()
	}
	return p
}