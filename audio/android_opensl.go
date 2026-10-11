//go:build android && cgo

package audio

/*
#cgo LDFLAGS: -lOpenSLES -llog
#include <stdlib.h>
void *svcCreateCapture(int sampleRate, int bufSamples, unsigned long long h);
int svcStartCapture(void *p);
void svcDestroyCapture(void *p);
void *svcCreatePlayback(int sampleRate, int bufSamples, unsigned long long h);
int svcStartPlayback(void *p);
void svcDestroyPlayback(void *p);
*/
import "C"

import (
	"errors"
	"runtime/cgo"
	"sync"
	"unsafe"
)

var ErrAudioUnavailable = errors.New("audio backend not available on Android")

type AndroidCapture struct {
	mu         sync.Mutex
	ctx        unsafe.Pointer
	handle     cgo.Handle
	hasHandle  bool
	sampleRate int
	bufSamples int
	ch         chan []float32
	leftover   []float32
}

func NewAndroidCapture(sampleRate, bufferSize int, deviceName string) (*AndroidCapture, error) {
	return &AndroidCapture{
		sampleRate: sampleRate,
		bufSamples: bufferSize,
		ch:         make(chan []float32, 64),
	}, nil
}

func (c *AndroidCapture) Open(sampleRate, bufferSize int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sampleRate = sampleRate
	c.bufSamples = bufferSize
	c.ch = make(chan []float32, 64)
	c.leftover = nil

	c.handle = cgo.NewHandle(c)
	c.hasHandle = true
	ctx := C.svcCreateCapture(C.int(sampleRate), C.int(bufferSize), C.ulonglong(c.handle))
	if ctx == nil {
		c.handle.Delete()
		c.hasHandle = false
		return ErrAudioUnavailable
	}
	c.ctx = ctx
	if C.svcStartCapture(ctx) != 0 {
		C.svcDestroyCapture(ctx)
		c.ctx = nil
		c.handle.Delete()
		c.hasHandle = false
		return ErrAudioUnavailable
	}
	return nil
}

func (c *AndroidCapture) Read(out []float32) (int, error) {
	c.mu.Lock()
	if c.ctx == nil {
		c.mu.Unlock()
		return 0, nil
	}
	leftover := c.leftover
	c.mu.Unlock()

	n := 0
	for n < len(out) {
		if len(leftover) == 0 {
			select {
			case leftover = <-c.ch:
			default:
				c.mu.Lock()
				c.leftover = leftover
				c.mu.Unlock()
				return n, nil
			}
		}
		m := copy(out[n:], leftover)
		leftover = leftover[m:]
		n += m
	}
	c.mu.Lock()
	c.leftover = leftover
	c.mu.Unlock()
	return n, nil
}

func (c *AndroidCapture) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ctx != nil {
		C.svcDestroyCapture(c.ctx)
		c.ctx = nil
	}
	if c.hasHandle {
		c.handle.Delete()
		c.hasHandle = false
	}
}

type AndroidPlayback struct {
	mu         sync.Mutex
	ctx        unsafe.Pointer
	handle     cgo.Handle
	hasHandle  bool
	sampleRate int
	bufSamples int
	ch         chan []float32
}

func NewAndroidPlayback(sampleRate, bufferSize int, deviceName string) (*AndroidPlayback, error) {
	return &AndroidPlayback{
		sampleRate: sampleRate,
		bufSamples: bufferSize,
		ch:         make(chan []float32, 64),
	}, nil
}

func (p *AndroidPlayback) Open(sampleRate, bufferSize int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sampleRate = sampleRate
	p.bufSamples = bufferSize
	p.ch = make(chan []float32, 64)

	p.handle = cgo.NewHandle(p)
	p.hasHandle = true
	ctx := C.svcCreatePlayback(C.int(sampleRate), C.int(bufferSize), C.ulonglong(p.handle))
	if ctx == nil {
		p.handle.Delete()
		p.hasHandle = false
		return ErrAudioUnavailable
	}
	p.ctx = ctx
	if C.svcStartPlayback(ctx) != 0 {
		C.svcDestroyPlayback(ctx)
		p.ctx = nil
		p.handle.Delete()
		p.hasHandle = false
		return ErrAudioUnavailable
	}
	return nil
}

func (p *AndroidPlayback) Write(samples []float32) error {
	p.mu.Lock()
	if p.ctx == nil {
		p.mu.Unlock()
		return nil
	}
	p.mu.Unlock()

	block := make([]float32, len(samples))
	copy(block, samples)
	select {
	case p.ch <- block:
	default:
	}
	return nil
}

func (p *AndroidPlayback) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ctx != nil {
		C.svcDestroyPlayback(p.ctx)
		p.ctx = nil
	}
	if p.hasHandle {
		p.handle.Delete()
		p.hasHandle = false
	}
}

//export goCapCallback
func goCapCallback(handle C.ulonglong, buf unsafe.Pointer, nsamples C.int) {
	v := cgo.Handle(uintptr(handle)).Value()
	c, ok := v.(*AndroidCapture)
	if !ok || c == nil {
		return
	}
	n := int(nsamples)
	src := unsafe.Slice((*int16)(buf), n)
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		out[i] = float32(src[i]) / 32768.0
	}
	select {
	case c.ch <- out:
	default:
	}
}

//export goPlayCallback
func goPlayCallback(handle C.ulonglong, buf unsafe.Pointer, nsamples C.int) {
	v := cgo.Handle(uintptr(handle)).Value()
	p, ok := v.(*AndroidPlayback)
	if !ok || p == nil {
		return
	}
	n := int(nsamples)
	dst := unsafe.Slice((*int16)(buf), n)

	var block []float32
	select {
	case block = <-p.ch:
	default:
	}
	for i := 0; i < n; i++ {
		var s float32
		if i < len(block) {
			s = block[i]
		}
		v := s * 32767.0
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		dst[i] = int16(v)
	}
}
