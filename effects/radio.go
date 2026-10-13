package effects

import "math"

// RadioEffect emulates an AM radio / telephone channel: a 250-3800 Hz
// band-pass, light overdrive and band-limited static whose level grows with
// strength.
type RadioEffect struct {
	BaseEffect
	hp     float64
	lp     float64
	prevIn float64
	seed   uint32
}

func NewRadioEffect() *RadioEffect {
	return &RadioEffect{
		BaseEffect: BaseEffect{
			name:     "radio",
			strength: 0,
			category: "modulation",
		},
		seed: 0x9E3779B9,
	}
}

// nextNoise advances a deterministic LCG returning white noise in [-1,1].
func (r *RadioEffect) nextNoise() float64 {
	r.seed = r.seed*1664525 + 1013904223
	return (float64(r.seed>>8)/float64(1<<23))*2 - 1
}

func (r *RadioEffect) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.85 * norm
	drive := 1.0 + 3.0*norm
	noiseGain := 0.09 * norm
	aHP := 1.0 / (1.0 + 48000.0/(2.0*math.Pi*250.0))
	aLP := 1.0 / (1.0 + 48000.0/(2.0*math.Pi*3800.0))
	for i, s := range samples {
		wide := float64(s)
		hpY := aHP * (r.hp + wide - r.prevIn)
		r.hp = hpY
		r.prevIn = wide
		lpY := r.lp + aLP*(hpY-r.lp)
		r.lp = lpY
		band := math.Tanh(lpY*drive + r.nextNoise()*noiseGain)
		samples[i] = float32(wide*(1-mix) + band*mix)
	}
}
