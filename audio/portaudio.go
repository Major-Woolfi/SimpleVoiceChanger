//go:build portaudio

package audio

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
)

// PortAudioCapture implements CaptureSource using PortAudio.
type PortAudioCapture struct {
	stream      *portaudio.Stream
	buffer      []float32
	bufferSize  int
	sampleRate  int
	mu          sync.Mutex
	closed      int32
	readPos     int
	writePos    int
	overflow    int32
}

// NewPortAudioCapture creates a new PortAudio capture source.
func NewPortAudioCapture(sampleRate, bufferSize int) (*PortAudioCapture, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, fmt.Errorf("portaudio init: %w", err)
	}

	c := &PortAudioCapture{
		bufferSize: bufferSize,
		sampleRate: sampleRate,
		buffer:     make([]float32, bufferSize*4),
	}

	return c, nil
}

// Open opens the capture stream.
func (c *PortAudioCapture) Open(sampleRate, bufferSize int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.closed) == 1 {
		return fmt.Errorf("capture already closed")
	}

	c.sampleRate = sampleRate
	c.bufferSize = bufferSize
	c.buffer = make([]float32, bufferSize*4)
	c.readPos = 0
	c.writePos = 0
	atomic.StoreInt32(&c.overflow, 0)

	device, err := portaudio.DefaultInputDevice()
	if err != nil {
		return fmt.Errorf("no input device: %w", err)
	}

	params := portaudio.LowLatencyParameters(device, nil)
	params.Input.Channels = 1
	params.SampleRate = float64(sampleRate)
	params.FramesPerBuffer = bufferSize

	stream, err := portaudio.OpenStream(params, c.processAudio)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}

	c.stream = stream
	err = stream.Start()
	if err != nil {
		stream.Close()
		c.stream = nil
		return fmt.Errorf("start stream: %w", err)
	}

	return nil
}

// processAudio is the PortAudio callback for capture.
func (c *PortAudioCapture) processAudio(in []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.closed) == 1 {
		return
	}

	for _, sample := range in {
		c.buffer[c.writePos] = sample
		c.writePos = (c.writePos + 1) % len(c.buffer)
		if c.writePos == c.readPos {
			atomic.StoreInt32(&c.overflow, 1)
			c.readPos = (c.readPos + 1) % len(c.buffer)
		}
	}
}

// Read reads audio samples from the capture buffer.
func (c *PortAudioCapture) Read(out []float32) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.closed) == 1 {
		return 0, fmt.Errorf("capture closed")
	}

	n := 0
	for i := 0; i < len(out); i++ {
		if c.readPos == c.writePos {
			break
		}
		out[i] = c.buffer[c.readPos]
		c.readPos = (c.readPos + 1) % len(c.buffer)
		n++
	}

	return n, nil
}

// Close closes the capture stream.
func (c *PortAudioCapture) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.SwapInt32(&c.closed, 1) == 1 {
		return
	}

	if c.stream != nil {
		c.stream.Stop()
		c.stream.Close()
		c.stream = nil
	}

	portaudio.Terminate()
}

// PortAudioPlayback implements PlaybackSink using PortAudio.
type PortAudioPlayback struct {
	stream      *portaudio.Stream
	buffer      []float32
	bufferSize  int
	sampleRate  int
	mu          sync.Mutex
	closed      int32
	readPos     int
	writePos    int
	underflow   int32
}

// NewPortAudioPlayback creates a new PortAudio playback sink.
func NewPortAudioPlayback(sampleRate, bufferSize int) (*PortAudioPlayback, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, fmt.Errorf("portaudio init: %w", err)
	}

	return &PortAudioPlayback{
		bufferSize: bufferSize,
		sampleRate: sampleRate,
		buffer:     make([]float32, bufferSize*4),
	}, nil
}

// Open opens the playback stream.
func (p *PortAudioPlayback) Open(sampleRate, bufferSize int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt32(&p.closed) == 1 {
		return fmt.Errorf("playback already closed")
	}

	p.sampleRate = sampleRate
	p.bufferSize = bufferSize
	p.buffer = make([]float32, bufferSize*4)
	p.readPos = 0
	p.writePos = 0
	atomic.StoreInt32(&p.underflow, 0)
	atomic.StoreInt32(&p.closed, 0)

	device, err := portaudio.DefaultOutputDevice()
	if err != nil {
		return fmt.Errorf("no output device: %w", err)
	}

	params := portaudio.LowLatencyParameters(nil, device)
	params.Output.Channels = 1
	params.SampleRate = float64(sampleRate)
	params.FramesPerBuffer = bufferSize

	stream, err := portaudio.OpenStream(params, p.processAudio)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}

	p.stream = stream
	err = stream.Start()
	if err != nil {
		stream.Close()
		p.stream = nil
		return fmt.Errorf("start stream: %w", err)
	}

	return nil
}

// processAudio is the PortAudio callback for playback.
func (p *PortAudioPlayback) processAudio(out []float32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt32(&p.closed) == 1 {
		for i := range out {
			out[i] = 0
		}
		return
	}

	for i := range out {
		if p.readPos == p.writePos {
			atomic.StoreInt32(&p.underflow, 1)
			out[i] = 0
		} else {
			out[i] = p.buffer[p.readPos]
			p.readPos = (p.readPos + 1) % len(p.buffer)
		}
	}
}

// Write writes audio samples to the playback buffer.
func (p *PortAudioPlayback) Write(samples []float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt32(&p.closed) == 1 {
		return fmt.Errorf("playback closed")
	}

	for _, s := range samples {
		p.buffer[p.writePos] = s
		p.writePos = (p.writePos + 1) % len(p.buffer)
		if p.writePos == p.readPos {
			p.readPos = (p.readPos + 1) % len(p.buffer)
		}
	}

	return nil
}

// Close closes the playback stream.
func (p *PortAudioPlayback) Close() {
	if atomic.SwapInt32(&p.closed, 1) == 1 {
		return
	}

	if p.stream != nil {
		p.stream.Stop()
		p.stream.Close()
		p.stream = nil
	}

	portaudio.Terminate()
}