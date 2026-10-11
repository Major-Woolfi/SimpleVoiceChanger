//go:build android && !cgo

package audio

import "errors"

var ErrAudioUnavailable = errors.New("audio backend not available on Android (build with cgo/NDK)")

type AndroidCapture struct{}

func NewAndroidCapture(sampleRate, bufferSize int, deviceName string) (*AndroidCapture, error) {
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

func NewAndroidPlayback(sampleRate, bufferSize int, deviceName string) (*AndroidPlayback, error) {
	return &AndroidPlayback{}, nil
}

func (p *AndroidPlayback) Open(sampleRate, bufferSize int) error {
	return ErrAudioUnavailable
}

func (p *AndroidPlayback) Write(samples []float32) error {
	return ErrAudioUnavailable
}

func (p *AndroidPlayback) Close() {}
