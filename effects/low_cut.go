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
	rate := 48000.0
	rc := 1.0 / (2.0 * math.Pi * l.cutoff * strength)
	dt := 1.0 / rate
	alpha := dt / (rc + dt)
	for i, s := range samples {
		filtered := float32(float64(l.state) + alpha*(float64(s)-float64(l.state)))
		mix := strength
		samples[i] = float32(float64(s)*(1.0-mix) + float64(filtered)*mix)
		l.state = filtered
	}
}
