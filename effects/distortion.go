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

// Process applies a tanh waveshaper whose drive and wet/dry mix scale
// linearly with strength, so low values give warm breakup instead of turning
// the voice into a square wave at once.
func (d *Distortion) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	drive := 1.0 + 9.0*norm
	mix := 0.35 + 0.65*norm
	for i, s := range samples {
		shaped := math.Tanh(float64(s) * drive)
		samples[i] = float32(float64(s)*(1.0-mix) + shaped*mix)
	}
}
