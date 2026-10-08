package audio

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

// Engine is the real-time audio pipeline: capture → noise layer → effect layer → playback.
type Engine struct {
	running     int32
	sampleRate  int
	bufferSize  int
	input       CaptureSource
	output      PlaybackSink
	pipeline    *Pipeline
	effectStren map[string]float64
	mu          sync.RWMutex
	stopCh      chan struct{}
	doneCh      chan struct{}
	stopped     int32
}

// Pipeline separates noise-reduction effects from the rest.
type Pipeline struct {
	noiseLayer  []effects.Effect
	effectLayer []effects.Effect
}

// NewEngine creates an engine with empty effect layers.
func NewEngine(sampleRate, bufferSize int) *Engine {
	return &Engine{
		sampleRate:  sampleRate,
		bufferSize:  bufferSize,
		effectStren: make(map[string]float64),
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		pipeline: &Pipeline{
			noiseLayer:  []effects.Effect{},
			effectLayer: []effects.Effect{},
		},
	}
}

// SetInput assigns the capture source.
func (e *Engine) SetInput(input CaptureSource) {
	e.input = input
}

// SetOutput assigns the playback sink.
func (e *Engine) SetOutput(output PlaybackSink) {
	e.output = output
}

// SetEffectStrength sets the strength (0-100) for an effect by key.
func (e *Engine) SetEffectStrength(name string, strength float64) {
	e.mu.Lock()
	e.effectStren[name] = strength
	e.mu.Unlock()
}

// GetEffectStrength returns the current strength for an effect.
func (e *Engine) GetEffectStrength(name string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.effectStren[name]
}

// GetEffectStrengths returns a copy of all effect strengths.
func (e *Engine) GetEffectStrengths() map[string]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]float64, len(e.effectStren))
	for k, v := range e.effectStren {
		out[k] = v
	}
	return out
}

// RegisterNoiseEffect adds an effect to the noise-reduction layer.
func (e *Engine) RegisterNoiseEffect(ev effects.Effect) {
	e.pipeline.noiseLayer = append(e.pipeline.noiseLayer, ev)
}

// RegisterEffect adds an effect to the main effect layer.
func (e *Engine) RegisterEffect(ev effects.Effect) {
	e.pipeline.effectLayer = append(e.pipeline.effectLayer, ev)
}

// Start opens input/output and launches the processing loop.
func (e *Engine) Start() error {
	if atomic.LoadInt32(&e.running) == 1 {
		return nil
	}
	atomic.StoreInt32(&e.running, 1)
	atomic.StoreInt32(&e.stopped, 0)

	if err := e.input.Open(e.sampleRate, e.bufferSize); err != nil {
		atomic.StoreInt32(&e.running, 0)
		return err
	}
	if err := e.output.Open(e.sampleRate, e.bufferSize); err != nil {
		e.input.Close()
		atomic.StoreInt32(&e.running, 0)
		return err
	}

	go e.runLoop()
	return nil
}

// Stop halts the processing loop and closes input/output.
func (e *Engine) Stop() {
	if atomic.LoadInt32(&e.running) == 0 {
		return
	}
	if atomic.SwapInt32(&e.stopped, 1) == 1 {
		return
	}
	atomic.StoreInt32(&e.running, 0)
	close(e.stopCh)
	<-e.doneCh
	e.input.Close()
	e.output.Close()
}

// runLoop is the core processing goroutine.
func (e *Engine) runLoop() {
	defer close(e.doneCh)
	buf := make([]float32, e.bufferSize)
	// Use a ticker to avoid busy-looping when the capture source
	// returns no data. The ticker is created once and never leaked.
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		if atomic.LoadInt32(&e.running) == 0 {
			return
		}

		n, err := e.input.Read(buf)
		if err != nil || n == 0 {
			// Wait briefly, but remain responsive to Stop().
			select {
			case <-e.stopCh:
				return
			case <-ticker.C:
				continue
			}
		}
		samples := buf[:n]

		e.mu.RLock()
		stren := make(map[string]float64, len(e.effectStren))
		for k, v := range e.effectStren {
			stren[k] = v
		}
		e.mu.RUnlock()

		noiseActive := stren["antinoise"] > 0 || stren["noise_suppressor"] > 0
		if noiseActive {
			for _, eff := range e.pipeline.noiseLayer {
				s := stren[eff.Name()]
				if s > 0 {
					eff.Process(samples, s)
				}
			}
		}

		for _, eff := range e.pipeline.effectLayer {
			s := stren[eff.Name()]
			if s > 0 {
				eff.Process(samples, s)
			}
		}

		if err := e.output.Write(samples); err != nil {
			continue
		}
	}
}

// CaptureSource is the input side of the audio pipeline.
type CaptureSource interface {
	Open(sampleRate, bufferSize int) error
	Read(out []float32) (int, error)
	Close()
}

// PlaybackSink is the output side of the audio pipeline.
type PlaybackSink interface {
	Open(sampleRate, bufferSize int) error
	Write(samples []float32) error
	Close()
}
