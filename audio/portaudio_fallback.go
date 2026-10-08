//go:build !portaudio && !android

package audio

import (
	"errors"
	"fmt"
)

// ErrAudioUnavailable is returned when audio backend is not available.
var ErrAudioUnavailable = errors.New("audio backend not available (build with -tags portaudio)")

// NewPortAudioCapture returns an error as portaudio is not enabled.
func NewPortAudioCapture(sampleRate, bufferSize int) (*PortAudioCapture, error) {
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
func NewPortAudioPlayback(sampleRate, bufferSize int) (*PortAudioPlayback, error) {
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

// NewDefaultCapture creates a mock capture source for testing.
func NewDefaultCapture() *DefaultCapture {
	return &DefaultCapture{
		data: make([]float32, 0),
	}
}

// DefaultCapture is a mock capture source for testing.
type DefaultCapture struct {
	data      []float32
	pos       int
	exhausted bool
}

func (d *DefaultCapture) Open(sampleRate, bufferSize int) error {
	return nil
}

func (d *DefaultCapture) Read(out []float32) (int, error) {
	if d.exhausted {
		return 0, fmt.Errorf("EOF")
	}
	n := 0
	for i := 0; i < len(out) && d.pos+i < len(d.data); i++ {
		out[i] = d.data[d.pos+i]
		n++
	}
	d.pos += n
	if n == 0 || d.pos >= len(d.data) {
		d.exhausted = true
	}
	return n, nil
}

func (d *DefaultCapture) Close() {}

// NewDefaultPlayback creates a mock playback sink for testing.
func NewDefaultPlayback() *DefaultPlayback {
	return &DefaultPlayback{
		written: make([]float32, 0),
	}
}

// DefaultPlayback is a mock playback sink for testing.
type DefaultPlayback struct {
	written []float32
}

func (d *DefaultPlayback) Open(sampleRate, bufferSize int) error {
	return nil
}

func (d *DefaultPlayback) Write(samples []float32) error {
	d.written = append(d.written, samples...)
	return nil
}

func (d *DefaultPlayback) Close() {}