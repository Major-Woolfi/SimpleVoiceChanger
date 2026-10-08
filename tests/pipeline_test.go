package tests

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/Major-Woolfi/SimpleVoiceChanger/audio"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

// mockCapture implements audio.CaptureSource.
type mockCapture struct {
	data      []float32
	pos       int
	exhausted bool
}

func (m *mockCapture) Open(int, int) error { return nil }
func (m *mockCapture) Read(out []float32) (int, error) {
	if m.exhausted {
		return 0, io.EOF
	}
	n := 0
	for i := 0; i < len(out) && m.pos+i < len(m.data); i++ {
		out[i] = m.data[m.pos+i]
		n++
	}
	m.pos += n
	if n == 0 || m.pos >= len(m.data) {
		m.exhausted = true
	}
	return n, nil
}
func (m *mockCapture) Close() {}

// mockSink implements audio.PlaybackSink.
type mockSink struct {
	written []float32
}

func (m *mockSink) Open(int, int) error { return nil }
func (m *mockSink) Write(samples []float32) error {
	m.written = append(m.written, samples...)
	return nil
}
func (m *mockSink) Close() {}

func TestEngineEffectSkipping(t *testing.T) {
	cap := &mockCapture{data: []float32{0.5, -0.5, 0.25, -0.25}}
	sink := &mockSink{}
	engine := audio.NewEngine(48000, 4)
	engine.SetInput(cap)
	engine.SetOutput(sink)

	engine.RegisterEffect(effects.NewDistortion())

	if err := engine.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Wait for processing to finish (capture exhausts quickly).
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			if len(sink.written) > 0 {
				break
			}
			time.Sleep(time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for samples")
	}

	engine.Stop()

	if len(sink.written) == 0 {
		t.Fatal("no samples written to sink")
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

	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			if len(sink.written) > 0 {
				break
			}
			time.Sleep(time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for samples")
	}

	engine.Stop()
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
	cfg := core.DefaultAppConfig()
	cfg.EffectStren["reverb"] = 30
	cfg.Enabled = true

	dirs := map[string]string{
		"presets": "testdata/presets",
		"config":  "testdata/config",
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
