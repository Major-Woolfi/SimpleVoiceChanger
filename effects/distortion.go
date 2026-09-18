package effects

import "math"

type Distortion struct {
	BaseEffect
}

func NewDistortion() *Distortion {
	return &Distortion{
		BaseEffect: BaseEffect{
			name:     "distortion",
			strength: 0,
			category: "modulation",
		},
	}
}

func (d *Distortion) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	drive := math.Tanh(strength * 4.0)
	for i, s := range samples {
		x := float64(s) * drive * (1.0 + strength*2.0)
		clipped := math.Tanh(x)
		samples[i] = float32(clipped)
	}
}
