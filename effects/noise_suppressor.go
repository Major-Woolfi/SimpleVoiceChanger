package effects

import "math"

type NoiseSuppressor struct {
	BaseEffect
	gateThreshold float64
}

func NewNoiseSuppressor() *NoiseSuppressor {
	return &NoiseSuppressor{
		BaseEffect: BaseEffect{
			name:     "noise_suppressor",
			strength: 0,
			category: "noise",
		},
		gateThreshold: 0.01,
	}
}

func (n *NoiseSuppressor) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	attack := 0.02
	release := 0.05
	var envelope float64
	for i, s := range samples {
		level := math.Abs(float64(s))
		target := 1.0
		if level < n.gateThreshold {
			target = float64(level/n.gateThreshold) * strength
			if target > 1.0 {
				target = 1.0
			}
		}
		if target > envelope {
			envelope += (target - envelope) * attack
		} else {
			envelope += (target - envelope) * release
		}
		if envelope < 0.001 {
			envelope = 0.001
		}
		samples[i] = float32(float64(s) * envelope)
	}
}
