package effects

import "math"

type RadioEffect struct {
	BaseEffect
	hp     float64
	lp     float64
	prevIn float64
}

func NewRadioEffect() *RadioEffect {
	return &RadioEffect{
		BaseEffect: BaseEffect{
			name:     "radio",
			strength: 0,
			category: "modulation",
		},
	}
}

// Process emulates an AM radio / telephone channel: a 300-3400 Hz band-pass
// followed by light overdrive. Strength controls how far the band is cut and
// how hard the drive hits.
func (r *RadioEffect) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.85 * norm
	drive := 1.0 + 3.0*norm
	// One-pole band-pass coefficients at 48 kHz. The band is wide enough to
	// keep the voice body so higher strengths still sound like a radio, not
	// silence.
	aHP := 1.0 / (1.0 + 48000.0/(2.0*math.Pi*250.0))
	aLP := 1.0 / (1.0 + 48000.0/(2.0*math.Pi*3800.0))
	for i, s := range samples {
		wide := float64(s)
		hpY := aHP * (r.hp + wide - r.prevIn)
		r.hp = hpY
		r.prevIn = wide
		lpY := r.lp + aLP*(hpY-r.lp)
		r.lp = lpY
		band := math.Tanh(lpY * drive)
		samples[i] = float32(wide*(1.0-mix) + band*mix)
	}
}
