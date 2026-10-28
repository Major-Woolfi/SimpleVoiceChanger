package tests

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Major-Woolfi/SimpleVoiceChanger/audio"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

// mockCapture implements audio.CaptureSource with an endlessly repeating
// buffer, like a live microphone.
type mockCapture struct {
	data []float32
	pos  int
}

func (m *mockCapture) Open(int, int) error { return nil }
func (m *mockCapture) Read(out []float32) (int, error) {
	n := 0
	for i := 0; i < len(out); i++ {
		out[i] = m.data[m.pos]
		m.pos = (m.pos + 1) % len(m.data)
		n++
	}
	return n, nil
}
func (m *mockCapture) Close() {}

// mockSink implements audio.PlaybackSink.
type mockSink struct {
	mu      sync.Mutex
	written []float32
}

func (m *mockSink) Open(int, int) error { return nil }
func (m *mockSink) Write(samples []float32) error {
	m.mu.Lock()
	m.written = append(m.written, samples...)
	m.mu.Unlock()
	return nil
}
func (m *mockSink) Close() {}

func (m *mockSink) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.written)
}

func waitForSamples(t *testing.T, sink *mockSink) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sink.count() > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timeout waiting for samples")
}

func TestEngineEffectSkipping(t *testing.T) {
	cap := &mockCapture{data: []float32{0.5, -0.5, 0.25, -0.25}}
	sink := &mockSink{}
	engine := audio.NewEngine(48000, 4)
	engine.SetInput(cap)
	engine.SetOutput(sink)

	engine.RegisterEffect(effects.NewDistortion(48000))

	if err := engine.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForSamples(t, sink)
	engine.Stop()

	if sink.count() == 0 {
		t.Fatal("no samples written to sink")
	}
}

func TestEngineRestart(t *testing.T) {
	engine := audio.NewEngine(48000, 4)
	engine.SetInput(&mockCapture{data: []float32{0.5, -0.5, 0.25, -0.25}})
	engine.SetOutput(&mockSink{})
	engine.RegisterEffect(effects.NewDistortion(48000))

	if err := engine.Start(); err != nil {
		t.Fatalf("first start: %v", err)
	}
	engine.Stop()

	engine.SetInput(&mockCapture{data: []float32{0.1, -0.1, 0.2, -0.2}})
	sink := &mockSink{}
	engine.SetOutput(sink)

	if err := engine.Start(); err != nil {
		t.Fatalf("second start: %v", err)
	}
	waitForSamples(t, sink)
	engine.Stop()

	if sink.count() == 0 {
		t.Fatal("no samples after restart")
	}
}

func TestEngineNoiseLayerSkippedWhenZero(t *testing.T) {
	cap := &mockCapture{data: []float32{0.1, 0.2, 0.3, 0.4}}
	sink := &mockSink{}
	engine := audio.NewEngine(48000, 4)
	engine.SetInput(cap)
	engine.SetOutput(sink)

	engine.RegisterNoiseEffect(effects.NewAntiNoise(48000))

	if err := engine.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitForSamples(t, sink)
	engine.Stop()
}

func TestEngineMonitorDoesNotBlockMainOutput(t *testing.T) {
	cap := &mockCapture{data: []float32{0.5, -0.5, 0.25, -0.25}}
	sink := &mockSink{}
	engine := audio.NewEngine(48000, 4)
	engine.SetInput(cap)
	engine.SetOutput(sink)
	engine.SetMonitor(false)

	if err := engine.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Main output must receive processed audio even when self-listening is off.
	waitForSamples(t, sink)
	engine.Stop()
	if sink.count() == 0 {
		t.Fatal("main output got no samples with monitor off")
	}
}

func TestPresetJSON(t *testing.T) {
	p := core.EffectPreset{
		Name: "test",
		Effects: map[string]float64{
			"compressor": 50,
			"reverb":     25,
		},
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(data) == "" {
		t.Fatal("empty JSON output")
	}

	var unmarshaled core.EffectPreset
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if unmarshaled.Name != "test" {
		t.Errorf("name mismatch: %s", unmarshaled.Name)
	}
	if unmarshaled.Effects["compressor"] != 50 {
		t.Errorf("compressor mismatch: %f", unmarshaled.Effects["compressor"])
	}
}

func TestAppConfigSaveLoad(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultAppConfig()
	cfg.EffectStren["reverb"] = 30
	cfg.Enabled = true

	dirs := map[string]string{
		"presets": filepath.Join(tmp, "presets"),
		"config":  filepath.Join(tmp, "config"),
	}
	cfg.Init(dirs)

	if err := cfg.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded := core.DefaultAppConfig()
	loaded.Init(dirs)
	if err := loaded.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.EffectStren["reverb"] != 30 {
		t.Errorf("reverb not persisted: %f", loaded.EffectStren["reverb"])
	}
	if !loaded.Enabled {
		t.Errorf("enabled not persisted")
	}
}
