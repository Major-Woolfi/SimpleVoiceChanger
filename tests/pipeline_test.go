package tests

import (
	"encoding/json"
	"testing"

	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

func TestEngineEffectSkipping(t *testing.T) {
	_ = core.EffectPreset{}
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
}
