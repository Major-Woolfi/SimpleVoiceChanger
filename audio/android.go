//go:build android

package audio

import (
	"errors"
	"fmt"
)

var ErrAudioUnavailable = errors.New("audio backend not available on Android (use platform-specific implementation)")

type AndroidCapture struct{}

func NewAndroidCapture(sampleRate, bufferSize int) (*AndroidCapture, error) {
	return &AndroidCapture{}, nil
}

func (c *AndroidCapture) Open(sampleRate, bufferSize int) error {
	return ErrAudioUnavailable
}

func (c *AndroidCapture) Read(out []float32) (int, error) {
	return 0, ErrAudioUnavailable
}

func (c *AndroidCapture) Close() {}

type AndroidPlayback struct{}

func NewAndroidPlayback(sampleRate, bufferSize int) (*AndroidPlayback, error) {
	return &AndroidPlayback{}, nil
}

func (p *AndroidPlayback) Open(sampleRate, bufferSize int) error {
	return ErrAudioUnavailable
}

func (p *AndroidPlayback) Write(samples []float32) error {
	return ErrAudioUnavailable
}

func (p *AndroidPlayback) Close() {}

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