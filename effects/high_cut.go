package effects

import "math"

type HighCut struct {
	BaseEffect
	cutoff float64
	state  float32
}

func NewHighCut() *HighCut {
	return &HighCut{
		BaseEffect: BaseEffect{
			name:     "high_cut",
			strength: 0,
			category: "filter",
		},
		cutoff: 16000.0,
	}
}

func (h *HighCut) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	cut := 24000.0 - 19000.0*norm
	rate := 48000.0
	rc := 1.0 / (2.0 * math.Pi * cut)
	alpha := 1.0 / (rc*rate + 1.0)
	mix := norm
	for i, s := range samples {
		filtered := float32(float64(h.state) + alpha*(float64(s)-float64(h.state)))
		h.state = filtered
		samples[i] = float32(float64(s)*(1.0-mix) + float64(filtered)*mix)
	}
}
