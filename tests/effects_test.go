package tests

import (
	"math"
	"testing"

	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

func TestAntiNoiseZeroStrength(t *testing.T) {
	e := effects.NewAntiNoise(48000)
	samples := []float32{0.1, 0.2, -0.1, -0.3}
	orig := append([]float32{}, samples...)
	e.Process(samples, 0)
	for i, s := range samples {
		if s != orig[i] {
			t.Errorf("sample %d changed with zero strength: %f", i, s)
		}
	}
}

func TestNoiseSuppressorZeroStrength(t *testing.T) {
	e := effects.NewNoiseSuppressor()
	samples := []float32{0.5, 0.3, -0.2}
	orig := append([]float32{}, samples...)
	e.Process(samples, 0)
	for i, s := range samples {
		if s != orig[i] {
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

func TestEffectNamesUnique(t *testing.T) {
	registry := effects.DefaultEffects()
	seen := map[string]bool{}
	for _, e := range registry {
		if seen[e.Key] {
			t.Errorf("duplicate effect key: %s", e.Key)
		}
		seen[e.Key] = true
	}
}

func TestEffectCategories(t *testing.T) {
	registry := effects.DefaultEffects()
	valid := map[string]bool{"noise": true, "filter": true, "modulation": true, "dynamics": true, "spatial": true, "voices": true}
	for _, e := range registry {
		if !valid[e.Category] {
			t.Errorf("effect %s has invalid category %s", e.Key, e.Category)
		}
	}
}

func TestAllEffectsProcessZero(t *testing.T) {
	samples := []float32{0.1, -0.2, 0.3, -0.4, 0.5}
	orig := append([]float32{}, samples...)
	constructors := []func() effects.Effect{
		func() effects.Effect { return effects.NewAntiNoise(48000) },
		func() effects.Effect { return effects.NewNoiseSuppressor() },
		func() effects.Effect { return effects.NewLowCut() },
		func() effects.Effect { return effects.NewHighCut() },
		func() effects.Effect { return effects.NewFormant() },
		func() effects.Effect { return effects.NewCompressor() },
		func() effects.Effect { return effects.NewRadioEffect() },
		func() effects.Effect { return effects.NewReverb() },
		func() effects.Effect { return effects.NewDistortion() },
		func() effects.Effect { return effects.NewSaturation() },
		func() effects.Effect { return effects.NewHarmonyEngine() },
		func() effects.Effect { return effects.NewChorus() },
		func() effects.Effect { return effects.NewFlanger() },
		func() effects.Effect { return effects.NewPhaser() },
		func() effects.Effect { return effects.NewDoubler() },
	}
	for _, c := range constructors {
		e := c()
		s := append([]float32{}, samples...)
		e.Process(s, 0)
		for i := range s {
			if math.Abs(float64(s[i]-orig[i])) > 1e-9 {
				t.Errorf("%s changed sample at zero strength", e.Name())
				break
			}
		}
	}
}

func TestAllEffectsProcessMax(t *testing.T) {
	samples := make([]float32, 1024)
	for i := range samples {
		samples[i] = 0.5
	}
	constructors := []func() effects.Effect{
		func() effects.Effect { return effects.NewAntiNoise(48000) },
		func() effects.Effect { return effects.NewNoiseSuppressor() },
		func() effects.Effect { return effects.NewLowCut() },
		func() effects.Effect { return effects.NewHighCut() },
		func() effects.Effect { return effects.NewFormant() },
		func() effects.Effect { return effects.NewCompressor() },
		func() effects.Effect { return effects.NewRadioEffect() },
		func() effects.Effect { return effects.NewReverb() },
		func() effects.Effect { return effects.NewDistortion() },
		func() effects.Effect { return effects.NewSaturation() },
		func() effects.Effect { return effects.NewHarmonyEngine() },
		func() effects.Effect { return effects.NewChorus() },
		func() effects.Effect { return effects.NewFlanger() },
		func() effects.Effect { return effects.NewPhaser() },
		func() effects.Effect { return effects.NewDoubler() },
	}
	for _, c := range constructors {
		e := c()
		s := append([]float32{}, samples...)
		e.Process(s, 100)
		// Ensure no NaN/Inf.
		for i, v := range s {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				t.Errorf("%s produced NaN/Inf at index %d", e.Name(), i)
			}
		}
	}
}
