package effects

import "math"

type LowCut struct {
	BaseEffect
	cutoff float64
	state  float32
}

func NewLowCut() *LowCut {
	return &LowCut{
		BaseEffect: BaseEffect{
			name:     "low_cut",
			strength: 0,
			category: "filter",
		},
		cutoff: 80.0,
	}
}

func (l *LowCut) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	cut := 40.0 + 420.0*norm
	rate := 48000.0
	rc := 1.0 / (2.0 * math.Pi * cut)
	alpha := 1.0 / (rc*rate + 1.0)
	mix := norm
	for i, s := range samples {
		low := float32(float64(l.state) + alpha*(float64(s)-float64(l.state)))
		l.state = low
		high := float64(s) - float64(low)
		samples[i] = float32(float64(s)*(1.0-mix) + high*mix)
	}
}
