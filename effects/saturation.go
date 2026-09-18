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

func (s *Saturation) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	drive := 2.0 + strength*8.0
	for i, sample := range samples {
		x := float64(sample) * drive
		var saturated float64
		if x > 0 {
			saturated = (1 - math.Exp(-x)) / (1 - math.Exp(-drive))
		} else {
			saturated = -(1 - math.Exp(x)) / (1 - math.Exp(-drive))
		}
		samples[i] = float32(saturated)
	}
}
