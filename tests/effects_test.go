package tests

import (
	"testing"

	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

func TestAntiNoiseZeroStrength(t *testing.T) {
	e := effects.NewAntiNoise(48000)
	samples := []float32{0.1, 0.2, -0.1, -0.3}
	e.Process(samples, 0)
	for i, s := range samples {
		if s != 0.1 && s != 0.2 && s != -0.1 && s != -0.3 {
			t.Errorf("sample %d changed with zero strength: %f", i, s)
		}
	}
}

func TestNoiseSuppressorZeroStrength(t *testing.T) {
	e := effects.NewNoiseSuppressor()
	samples := []float32{0.5, 0.3, -0.2}
	e.Process(samples, 0)
	for i, s := range samples {
		if s != 0.5 && s != 0.3 && s != -0.2 {
			t.Errorf("sample %d changed with zero strength: %f", i, s)
		}
	}
}

func TestEffectStrengthBounds(t *testing.T) {
	e := effects.NewDistortion()
	e.SetStrength(-10)
	if e.Strength() != 0 {
		t.Errorf("negative strength not clamped: %f", e.Strength())
	}
	e.SetStrength(150)
	if e.Strength() != 100 {
		t.Errorf("strength above 100 not clamped: %f", e.Strength())
	}
	e.SetStrength(50)
	if e.Strength() != 50 {
		t.Errorf("strength not set correctly: %f", e.Strength())
	}
}
