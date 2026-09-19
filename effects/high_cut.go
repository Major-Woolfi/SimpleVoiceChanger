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
	rate := 48000.0
	rc := 1.0 / (2.0 * math.Pi * h.cutoff * strength)
	dt := 1.0 / rate
	alpha := dt / (rc + dt)
	for i, s := range samples {
		filtered := float32(float64(h.state) + alpha*(float64(s)-float64(h.state)))
		mix := strength
		samples[i] = float32(float64(s)*(1.0-mix) + float64(filtered)*mix)
		h.state = filtered
	}
}
