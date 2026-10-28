package effects

import "math"

type Distortion struct {
	BaseEffect
	sampleRate int
	lp         float64 // stateful lowpass output for anti-aliasing
}

func NewDistortion(sampleRate int) *Distortion {
	return &Distortion{
		BaseEffect: BaseEffect{
			name:     "distortion",
			strength: 0,
			category: "modulation",
		},
		sampleRate: sampleRate,
	}
}

// Process applies a knee soft-clipper: unity gain below knee, smoothly
// compresses above. A stateful lowpass filter tames high-frequency products
// so the effect never sounds like a volume explosion.
func (d *Distortion) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	knee := 0.35 - 0.23*norm
	drive := 1.0 + 10.0*norm
	mix := norm
	// Alpha scales with sample rate so the filter cutoff stays proportional.
	coeff := 0.15 + 0.15*float64(d.sampleRate)/48000.0
	for i, s := range samples {
		x := drive * float64(s)
		var clipped float64
		if x > knee {
			clipped = knee + (1.0-knee)*math.Tanh((x-knee)/(1.0-knee))
		} else if x < -knee {
			clipped = -knee - (1.0-knee)*math.Tanh((-x-knee)/(1.0-knee))
		} else {
			clipped = x
		}
		d.lp = d.lp*(1.0-coeff) + clipped*coeff
		samples[i] = float32(float64(s)*(1.0-mix) + d.lp*mix)
	}
}
