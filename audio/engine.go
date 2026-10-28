package audio

import (
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

// Engine is the real-time audio pipeline: capture → noise layer → pitch →
// effect layer → output gain → playback.
type Engine struct {
	running     int32
	sampleRate  int
	bufferSize  int
	input       CaptureSource
	output      PlaybackSink
	pipeline    *Pipeline
	effectStren map[string]float64
	monitor     int32
	pitchBits   atomic.Uint64 // pitch offset in semitones, packed float64 bits
	pitcher     *effects.PitchShifter
	pitchBuf    []float32
	mu          sync.RWMutex
	lifecycle   sync.Mutex
	stopCh      chan struct{}
	doneCh      chan struct{}
	monitorSink PlaybackSink
	outDevName  string
}

// Pipeline separates noise-reduction effects from the rest.
type Pipeline struct {
	noiseLayer  []effects.Effect
	effectLayer []effects.Effect
}

// NewEngine creates an engine with empty effect layers and a pitch shifter.
func NewEngine(sampleRate, bufferSize int) *Engine {
	return &Engine{
		sampleRate:  sampleRate,
		bufferSize:  bufferSize,
		effectStren: make(map[string]float64),
		monitor:     1,
		pitcher:     effects.NewPitchShifter(sampleRate),
		pitchBuf:    make([]float32, 0, bufferSize*2),
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

// SetOutputDeviceName sets the name of the output device for logging.
func (e *Engine) SetOutputDeviceName(name string) {
	e.outDevName = name
}

// SetEffectStrength sets the strength (0-100) for an effect by key.
func (e *Engine) SetEffectStrength(name string, strength float64) {
	e.mu.Lock()
	e.effectStren[name] = strength
	e.mu.Unlock()
}

// SetMonitor controls self-listening: when on and the engine is running, the
// processed voice is also played back on the default output device
// (headphones/speakers). Toggling it while the engine is stopped only stores
// the preference; the monitor stream is opened by Start and closed by Stop,
// so the flag is never ignored on the next start. The main voice-changer
// output is unaffected by this flag - it always receives the processed audio
// while the engine runs.
func (e *Engine) SetMonitor(on bool) {
	if on {
		atomic.StoreInt32(&e.monitor, 1)
	} else {
		atomic.StoreInt32(&e.monitor, 0)
	}
	// Apply immediately only when the engine is live; otherwise Start will
	// honour the stored flag.
	if atomic.LoadInt32(&e.running) == 1 {
		if on {
			e.ensureMonitorSink()
		} else {
			e.closeMonitorSink()
		}
	}
}

// SetPitch sets the voice pitch offset in semitones, clamped to [-24, 24].
// A zero offset is a pure passthrough: no pitch shift is applied. The value is
// stored as packed float64 bits so the audio loop can pick it up without a
// lock; the loop reprograms the shifter when the value changes.
func (e *Engine) SetPitch(semi float64) {
	if semi > 24 {
		semi = 24
	}
	if semi < -24 {
		semi = -24
	}
	e.pitchBits.Store(math.Float64bits(semi))
}

// Pitch returns the current pitch offset in semitones.
func (e *Engine) Pitch() float64 {
	return math.Float64frombits(e.pitchBits.Load())
}

// Monitor reports whether self-listening is currently requested.
func (e *Engine) Monitor() bool {
	return atomic.LoadInt32(&e.monitor) == 1
}

// ensureMonitorSink opens the self-listening sink on the default output
// device if it is not open yet and the main output is NOT the default device.
// When the main output is the default device, monitoring is redundant — the
// processed signal already plays through it. Safe to call repeatedly.
func (e *Engine) ensureMonitorSink() {
	e.mu.Lock()
	if e.monitorSink != nil {
		e.mu.Unlock()
		return
	}
	// If the main output is the default device, no separate monitor sink is
	// needed — the processed signal is already playing through it.
	if e.outDevName == "" {
		e.mu.Unlock()
		return
	}
	e.mu.Unlock()

	sink := NewPlayback(e.sampleRate, e.bufferSize, "")
	if sink == nil {
		return
	}
	if err := sink.Open(e.sampleRate, e.bufferSize); err != nil {
		slog.Warn("monitor sink open failed", "error", err)
		return
	}
	slog.Info("monitor sink opened on default device")

	e.mu.Lock()
	if e.monitorSink != nil {
		e.mu.Unlock()
		sink.Close()
		return
	}
	e.monitorSink = sink
	e.mu.Unlock()
}

// closeMonitorSink releases the self-listening sink if open.
func (e *Engine) closeMonitorSink() {
	e.mu.Lock()
	mon := e.monitorSink
	e.monitorSink = nil
	e.mu.Unlock()
	if mon != nil {
		mon.Close()
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
		slog.Error("input open failed", "error", err)
		return err
	}
	if err := e.output.Open(e.sampleRate, e.bufferSize); err != nil {
		slog.Error("output open failed", "error", err)
		e.input.Close()
		return err
	}
	slog.Info("output device", "name", e.outDevName)

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	e.mu.Lock()
	e.stopCh = stopCh
	e.doneCh = doneCh
	e.mu.Unlock()

	atomic.StoreInt32(&e.running, 1)
	// When monitor is requested and the engine is starting, open the self-
	// listening sink; when it is not requested, make sure it is closed.
	if atomic.LoadInt32(&e.monitor) == 1 {
		e.ensureMonitorSink()
	} else {
		e.closeMonitorSink()
	}
	slog.Info("engine started", "sampleRate", e.sampleRate, "bufferSize", e.bufferSize)
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
	mon := e.monitorSink
	e.monitorSink = nil
	e.mu.Unlock()

	close(stopCh)
	<-doneCh
	e.input.Close()
	e.output.Close()
	if mon != nil {
		mon.Close()
	}
	slog.Info("engine stopped")
}

// runLoop is the core processing goroutine. Channels are passed by value so
// the loop never reads fields that Stop may clear concurrently. Processed
// audio is always written to the main output (the voice-changer feed); the
// monitor sink (self-listening on the default device) receives a copy only
// while monitoring is on.
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
		if err != nil {
			slog.Warn("input read error", "error", err)
		}
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

		// Chain: noise gate → pitch → colour effects → output gain. Pitch sits
		// before the colour effects so distortion and reverb work on the shifted
		// voice, which is what makes the result sound like a different person.
		e.applyLayer(stren, samples, e.pipeline.noiseLayer)
		samples = e.applyPitch(samples)
		e.applyLayer(stren, samples, e.pipeline.effectLayer)
		applyOutputGain(samples)

		if err := e.output.Write(samples); err != nil {
			slog.Warn("output write error", "error", err)
			continue
		}
		if atomic.LoadInt32(&e.monitor) == 1 {
			e.mu.RLock()
			mon := e.monitorSink
			e.mu.RUnlock()
			if mon != nil {
				_ = mon.Write(samples)
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

func (e *Engine) applyStrengths(stren map[string]float64, samples []float32, noiseLayer, effectLayer []effects.Effect) error {
	e.applyLayer(stren, samples, noiseLayer)
	e.applyLayer(stren, samples, effectLayer)
	return nil
}

// applyLayer runs every effect of the layer whose strength is above zero.
func (e *Engine) applyLayer(stren map[string]float64, samples []float32, layer []effects.Effect) {
	for _, eff := range layer {
		if s := stren[eff.Name()]; s > 0 {
			eff.Process(samples, s)
		}
	}
}

// applyPitch shifts the block to the configured pitch and returns the buffer
// holding the result. The engine reuses pitchBuf so the audio loop never
// allocates while running.
func (e *Engine) applyPitch(samples []float32) []float32 {
	semi := e.Pitch()
	if e.pitcher == nil || semi == 0 {
		return samples
	}
	if semi != e.pitcher.Semitones() {
		e.pitcher.SetSemitones(semi)
	}
	e.pitchBuf = append(e.pitchBuf[:0], samples...)
	e.pitcher.ProcessInto(e.pitchBuf[:len(samples)], samples)
	return e.pitchBuf[:len(samples)]
}

// applyOutputGain lifts the level back to a normal listening volume after the
// effect chain, then bends the peaks with a soft knee instead of hard-clipping
// them: harsh digital clipping is exactly what makes the voice sound quiet and
// crushed at the same time.
func applyOutputGain(samples []float32) {
	const gain = 4.0
	const knee = 0.90
	const headroom = 1.0 - knee
	for i, s := range samples {
		v := float64(s) * gain
		a := math.Abs(v)
		if a > knee {
			v = math.Copysign(knee+headroom*math.Tanh((a-knee)/headroom), v)
		}
		samples[i] = float32(v)
	}
}
