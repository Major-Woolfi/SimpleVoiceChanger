package effects

import "math"

type Saturation struct {
	BaseEffect
}

func NewSaturation() *Saturation {
	return &Saturation{
		BaseEffect: BaseEffect{
			name:     "saturation",
			strength: 0,
			category: "modulation",
		},
	}
}

// Process applies an exponential waveshaper. The curve is almost linear for
// small inputs (warmth) and only compresses hard peaks, with the drive
// growing gradually with strength.
func (s *Saturation) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	drive := 1.0 + 6.0*norm
	mix := 0.4 + 0.6*norm
	denom := 1.0 - math.Exp(-drive)
	for i, sample := range samples {
		x := float64(sample) * drive
		var saturated float64
		if x > 0 {
			saturated = (1 - math.Exp(-x)) / denom
		} else {
			saturated = -(1 - math.Exp(x)) / denom
		}
		samples[i] = float32(float64(sample)*(1.0-mix) + saturated*mix)
	}
}
