package audio

import (
	"sync"
	"sync/atomic"

	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

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
}

type Pipeline struct {
	noiseLayer  []effects.Effect
	effectLayer []effects.Effect
}

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

func (e *Engine) SetInput(input CaptureSource) {
	e.input = input
}

func (e *Engine) SetOutput(output PlaybackSink) {
	e.output = output
}

func (e *Engine) SetEffectStrength(name string, strength float64) {
	e.mu.Lock()
	e.effectStren[name] = strength
	e.mu.Unlock()
}

func (e *Engine) GetEffectStrength(name string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.effectStren[name]
}

func (e *Engine) RegisterNoiseEffect(ev effects.Effect) {
	e.pipeline.noiseLayer = append(e.pipeline.noiseLayer, ev)
}

func (e *Engine) RegisterEffect(ev effects.Effect) {
	e.pipeline.effectLayer = append(e.pipeline.effectLayer, ev)
}

func (e *Engine) Start() error {
	if atomic.LoadInt32(&e.running) == 1 {
		return nil
	}
	atomic.StoreInt32(&e.running, 1)

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

func (e *Engine) Stop() {
	if atomic.LoadInt32(&e.running) == 0 {
		return
	}
	atomic.StoreInt32(&e.running, 0)
	close(e.stopCh)
	<-e.doneCh
	e.input.Close()
	e.output.Close()
}

func (e *Engine) runLoop() {
	defer close(e.doneCh)
	buf := make([]float32, e.bufferSize)

	for {
		if atomic.LoadInt32(&e.running) == 0 {
			return
		}

		n, err := e.input.Read(buf)
		if err != nil || n == 0 {
			continue
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

type CaptureSource interface {
	Open(sampleRate, bufferSize int) error
	Read(out []float32) (int, error)
	Close()
}

type PlaybackSink interface {
	Open(sampleRate, bufferSize int) error
	Write(samples []float32) error
	Close()
}
