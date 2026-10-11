//go:build !portaudio && !android

package audio

import "errors"

// ErrAudioUnavailable is returned when audio backend is not available.
var ErrAudioUnavailable = errors.New("audio backend not available (build with -tags portaudio)")

// NewPortAudioCapture returns an error as portaudio is not enabled.
func NewPortAudioCapture(sampleRate, bufferSize int, deviceName string) (*PortAudioCapture, error) {
	return nil, ErrAudioUnavailable
}

// PortAudioCapture is a stub when portaudio is not enabled.
type PortAudioCapture struct{}

func (c *PortAudioCapture) Open(sampleRate, bufferSize int) error {
	return ErrAudioUnavailable
}

func (c *PortAudioCapture) Read(out []float32) (int, error) {
	return 0, ErrAudioUnavailable
}

func (c *PortAudioCapture) Close() {}

// NewPortAudioPlayback returns an error as portaudio is not enabled.
func NewPortAudioPlayback(sampleRate, bufferSize int, deviceName string) (*PortAudioPlayback, error) {
	return nil, ErrAudioUnavailable
}

// PortAudioPlayback is a stub when portaudio is not enabled.
type PortAudioPlayback struct{}

func (p *PortAudioPlayback) Open(sampleRate, bufferSize int) error {
	return ErrAudioUnavailable
}

func (p *PortAudioPlayback) Write(samples []float32) error {
	return ErrAudioUnavailable
}

func (p *PortAudioPlayback) Close() {}
