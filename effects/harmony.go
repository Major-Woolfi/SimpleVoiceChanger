package effects

import "math"

type HarmonyEngine struct {
	BaseEffect
	phase float64
}

func NewHarmonyEngine() *HarmonyEngine {
	return &HarmonyEngine{
		BaseEffect: BaseEffect{
			name:     "harmony",
			strength: 0,
			category: "voices",
		},
	}
}

func (h *HarmonyEngine) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	mix := strength * 0.4
	pitches := []float64{1.0, 1.5, 2.0}
	for i, s := range samples {
		added := float64(s)
		for _, p := range pitches {
			if p == 1.0 {
				continue
			}
			phase := h.phase * p
			weight := math.Sin(phase) * mix * 0.3
			added += float64(s) * weight * strength * 0.2
		}
		h.phase += 0.01 * strength
		if h.phase > math.Pi*2 {
			h.phase -= math.Pi * 2
		}
		samples[i] = float32(added)
	}
}
