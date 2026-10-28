package effects

import "math"

type HighCut struct {
	BaseEffect
	sampleRate int
	state      float32
}

func NewHighCut(sampleRate int) *HighCut {
	return &HighCut{
		BaseEffect: BaseEffect{
			name:     "high_cut",
			strength: 0,
			category: "filter",
		},
		sampleRate: sampleRate,
	}
}

func (h *HighCut) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	// Nyquist-scaled cutoff so the filter works at any sample rate.
	maxHz := float64(h.sampleRate) * 0.45
	cut := maxHz - (maxHz-3000.0)*norm
	rc := 1.0 / (2.0 * math.Pi * cut)
	alpha := 1.0 / (rc*float64(h.sampleRate) + 1.0)
	mix := norm
	for i, s := range samples {
		filtered := float32(float64(h.state) + alpha*(float64(s)-float64(h.state)))
		h.state = filtered
		samples[i] = float32(float64(s)*(1.0-mix) + float64(filtered)*mix)
	}
}
