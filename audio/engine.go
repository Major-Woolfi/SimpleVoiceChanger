package audio

import (
	"fmt"
	"math"
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
	monitor     int32
	mu          sync.RWMutex
	lifecycle   sync.Mutex
	stopCh      chan struct{}
	doneCh      chan struct{}
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
		monitor:     1,
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

// SetMonitor controls whether the processed voice is written to the output
// device (you hear yourself). When monitoring is off the engine still reads
// the microphone and applies effects, but plays nothing back.
func (e *Engine) SetMonitor(on bool) {
	if on {
		atomic.StoreInt32(&e.monitor, 1)
	} else {
		atomic.StoreInt32(&e.monitor, 0)
	}
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
	e.lifecycle.Lock()
	defer e.lifecycle.Unlock()

	if atomic.LoadInt32(&e.running) == 1 {
		return nil
	}
	if err := e.input.Open(e.sampleRate, e.bufferSize); err != nil {
		return err
	}
	if err := e.output.Open(e.sampleRate, e.bufferSize); err != nil {
		e.input.Close()
		return err
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	e.mu.Lock()
	e.stopCh = stopCh
	e.doneCh = doneCh
	e.mu.Unlock()

	atomic.StoreInt32(&e.running, 1)
	go e.runLoop(stopCh, doneCh)
	return nil
}

// Stop halts the processing loop and closes input/output.
func (e *Engine) Stop() {
	e.lifecycle.Lock()
	defer e.lifecycle.Unlock()

	if atomic.LoadInt32(&e.running) == 0 {
		return
	}
	atomic.StoreInt32(&e.running, 0)

	e.mu.Lock()
	stopCh, doneCh := e.stopCh, e.doneCh
	e.stopCh = nil
	e.doneCh = nil
	e.mu.Unlock()

	close(stopCh)
	<-doneCh
	e.input.Close()
	e.output.Close()
}

// runLoop is the core processing goroutine. Channels are passed by value so
// the loop never reads fields that Stop may clear concurrently.
func (e *Engine) runLoop(stopCh, doneCh chan struct{}) {
	defer close(doneCh)
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
			case <-stopCh:
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

		_ = e.applyStrengths(stren, samples, e.pipeline.noiseLayer, e.pipeline.effectLayer)

		if atomic.LoadInt32(&e.monitor) == 1 {
			if err := e.output.Write(samples); err != nil {
				continue
			}
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

// SelfTest validates the effect chain by running a synthetic signal
// through every registered effect at full strength. It returns an error
// if any effect produces NaN, Inf or silence.
func (e *Engine) SelfTest() error {
	e.mu.RLock()
	stren := make(map[string]float64, len(e.effectStren))
	for k, v := range e.effectStren {
		stren[k] = v
	}
	noise := e.pipeline.noiseLayer
	effects := e.pipeline.effectLayer
	e.mu.RUnlock()

	buf := make([]float32, e.bufferSize)
	for i := range buf {
		buf[i] = float32(i%64)/64*0.8 - 0.4
	}
	if err := e.applyStrengths(stren, buf, noise, effects); err != nil {
		return err
	}
	for i, v := range buf {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return fmt.Errorf("NaN/Inf at sample %d", i)
		}
	}
	return nil
}

// PlayTest renders a short processed tone through the given output device so
// the user can verify the audio path. An empty deviceName uses the default.
func (e *Engine) PlayTest(deviceName string) error {
	if err := e.SelfTest(); err != nil {
		return err
	}

	sink := NewPlayback(e.sampleRate, e.bufferSize, deviceName)
	if sink == nil {
		return fmt.Errorf("no playback backend")
	}
	if err := sink.Open(e.sampleRate, e.bufferSize); err != nil {
		return err
	}
	defer sink.Close()

	e.mu.RLock()
	stren := make(map[string]float64, len(e.effectStren))
	for k, v := range e.effectStren {
		stren[k] = v
	}
	noise := e.pipeline.noiseLayer
	effectLayer := e.pipeline.effectLayer
	e.mu.RUnlock()

	total := e.sampleRate * 3 / 2
	buf := make([]float32, e.bufferSize)
	fadeSamples := e.sampleRate / 200 // 5ms fade in/out avoids clicks
	writeChunk := func(written, n int) error {
		for i := 0; i < n; i++ {
			idx := written + i
			t := float64(idx) / float64(e.sampleRate)
			buf[i] = float32(0.45 * math.Sin(2*math.Pi*880*t))
			if idx < fadeSamples {
				buf[i] = float32(float64(buf[i]) * float64(idx) / float64(fadeSamples))
			} else if idx >= total-fadeSamples {
				buf[i] = float32(float64(buf[i]) * float64(total-idx) / float64(fadeSamples))
			}
		}
		if err := e.applyStrengths(stren, buf[:n], noise, effectLayer); err != nil {
			return err
		}
		return sink.Write(buf[:n])
	}

	// When the sink exposes its buffered level, keep a few buffers queued so
	// scheduler jitter cannot starve the device (a per-buffer time.Sleep used
	// to underflow on Windows and produced the choppy/crackling test tone).
	if bs, ok := sink.(interface{ Buffered() int }); ok {
		high := e.bufferSize * 3
		deadline := time.Now().Add(10 * time.Second)
		for written := 0; written < total; {
			if time.Now().After(deadline) {
				return fmt.Errorf("playback stalled")
			}
			if bs.Buffered() >= high {
				time.Sleep(2 * time.Millisecond)
				continue
			}
			n := e.bufferSize
			if written+n > total {
				n = total - written
			}
			if err := writeChunk(written, n); err != nil {
				return err
			}
			written += n
		}
		for i := 0; i < 600 && bs.Buffered() > 0; i++ {
			time.Sleep(5 * time.Millisecond)
		}
		return nil
	}

	interval := time.Duration(float64(e.bufferSize) / float64(e.sampleRate) * float64(time.Second))
	for written := 0; written < total; written += e.bufferSize {
		n := e.bufferSize
		if written+n > total {
			n = total - written
		}
		if err := writeChunk(written, n); err != nil {
			return err
		}
		time.Sleep(interval)
	}
	return nil
}

// TestLive broadcasts the live microphone through the current effect setup to
// the given output device for a few seconds, so the user can hear their own
// processed voice on the speakers without changing normal monitoring.
func (e *Engine) TestLive(inputDevice, outputDevice string, seconds int) error {
	if seconds <= 0 {
		seconds = 6
	}
	capSrc := NewCapture(e.sampleRate, e.bufferSize, inputDevice)
	if capSrc == nil {
		return fmt.Errorf("no capture backend")
	}
	sink := NewPlayback(e.sampleRate, e.bufferSize, outputDevice)
	if sink == nil {
		return fmt.Errorf("no playback backend")
	}
	if err := capSrc.Open(e.sampleRate, e.bufferSize); err != nil {
		return err
	}
	defer capSrc.Close()
	if err := sink.Open(e.sampleRate, e.bufferSize); err != nil {
		return err
	}
	defer sink.Close()

	stren := make(map[string]float64, 15)
	noise, effectLayer := freshLayers(e.sampleRate)
	for _, eff := range noise {
		stren[eff.Name()] = e.GetEffectStrength(eff.Name())
	}
	for _, eff := range effectLayer {
		stren[eff.Name()] = e.GetEffectStrength(eff.Name())
	}

	total := e.sampleRate * seconds
	buf := make([]float32, e.bufferSize)
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	hasBuffered := false
	var bs interface{ Buffered() int }
	if sink2, ok := sink.(interface{ Buffered() int }); ok {
		hasBuffered = true
		bs = sink2
	}

	for written := 0; written < total; {
		n, err := capSrc.Read(buf)
		if err != nil || n == 0 {
			<-ticker.C
			continue
		}
		samples := buf[:n]
		if err := e.applyStrengths(stren, samples, noise, effectLayer); err != nil {
			return err
		}
		if hasBuffered {
			deadline := time.Now().Add(10 * time.Second)
			for bs.Buffered() >= e.bufferSize*3 {
				if time.Now().After(deadline) {
					return fmt.Errorf("playback stalled")
				}
				time.Sleep(2 * time.Millisecond)
			}
		}
		if err := sink.Write(samples); err != nil {
			return err
		}
		written += n
	}
	if hasBuffered {
		for i := 0; i < 600 && bs.Buffered() > 0; i++ {
			time.Sleep(5 * time.Millisecond)
		}
	}
	return nil
}

// freshLayers builds clean effect instances for transient tests so no state is
// shared with a running engine.
func freshLayers(sampleRate int) ([]effects.Effect, []effects.Effect) {
	noise := []effects.Effect{effects.NewAntiNoise(sampleRate), effects.NewNoiseSuppressor()}
	signal := []effects.Effect{
		effects.NewLowCut(),
		effects.NewHighCut(),
		effects.NewFormant(),
		effects.NewCompressor(),
		effects.NewRadioEffect(),
		effects.NewReverb(),
		effects.NewDistortion(),
		effects.NewSaturation(),
		effects.NewHarmonyEngine(),
		effects.NewChorus(),
		effects.NewFlanger(),
		effects.NewPhaser(),
		effects.NewDoubler(),
	}
	return noise, signal
}

func (e *Engine) applyStrengths(stren map[string]float64, samples []float32, noiseLayer, effectLayer []effects.Effect) error {
	noiseActive := stren["antinoise"] > 0 || stren["noise_suppressor"] > 0
	if noiseActive {
		for _, eff := range noiseLayer {
			if s := stren[eff.Name()]; s > 0 {
				eff.Process(samples, s)
			}
		}
	}
	for _, eff := range effectLayer {
		if s := stren[eff.Name()]; s > 0 {
			eff.Process(samples, s)
		}
	}
	return nil
}
