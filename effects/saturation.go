package effects

import "math"

type Saturation struct {
	BaseEffect
	sampleRate int
}

func NewSaturation(sampleRate int) *Saturation {
	return &Saturation{
		BaseEffect: BaseEffect{
			name:     "saturation",
			strength: 0,
			category: "modulation",
		},
		sampleRate: sampleRate,
	}
}

// Process applies a knee soft-clipper: unity gain for signals below knee,
// gentle compression for peaks. The effect sounds like tube warm-up: the
// voice gets thicker without losing volume.
func (s *Saturation) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	knee := 0.75 - 0.50*norm // 0.75 → 0.25
	drive := 1.0 + 1.5*norm  // mild pre-gain
	mix := norm              // dry at the minimum, fully wet at the maximum
	for i, x := range samples {
		sx := float64(x) * drive
		var out float64
		if sx > knee {
			out = knee + (1.0-knee)*math.Tanh((sx-knee)/(1.0-knee))
		} else if sx < -knee {
			out = -knee - (1.0-knee)*math.Tanh((-sx-knee)/(1.0-knee))
		} else {
			out = sx
		}
		samples[i] = float32(float64(x)*(1.0-mix) + out*mix)
	}
}
