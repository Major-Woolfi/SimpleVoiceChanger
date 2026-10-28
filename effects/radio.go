package effects

import "math"

// RadioEffect emulates an AM radio / telephone channel: a steep 300-3500 Hz
// band-pass, hard clipping, band-limited static and a slow fading wobble.
type RadioEffect struct {
	BaseEffect
	sampleRate int
	hp1, hp2   float64 // 2nd-order HPF state
	lp1, lp2   float64 // 2nd-order LPF state
	prevIn     float64
	seed       uint32
	fadePhase  float64 // slow AM fading wobble
}

func NewRadioEffect(sampleRate int) *RadioEffect {
	return &RadioEffect{
		BaseEffect: BaseEffect{
			name:     "radio",
			strength: 0,
			category: "modulation",
		},
		sampleRate: sampleRate,
		seed:       0x9E3779B9,
	}
}

// nextNoise advances a deterministic LCG returning white noise in [-1,1].
func (r *RadioEffect) nextNoise() float64 {
	r.seed = r.seed*1664525 + 1013904223
	return (float64(r.seed>>8)/float64(1<<23))*2 - 1
}

// Process emulates an AM radio / telephone channel: a steep band-pass,
// hard clipping, radio static, and a slow fading wobble. The wet mix grows
// from zero with the strength, so the minimum setting is a light tint instead
// of a full radio channel.
func (r *RadioEffect) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	drive := 2.0 + 6.0*norm
	noiseGain := 0.10 * norm
	mix := norm
	// 2nd-order HPF at 300 Hz and LPF at 3500 Hz for the steep radio band.
	fs := float64(r.sampleRate)
	wcHP := 2 * math.Pi * 300.0 / fs
	aHP := wcHP / (1.0 + wcHP)
	wcLP := 2 * math.Pi * 3500.0 / fs
	aLP := wcLP / (1.0 + wcLP)
	// Slow AM fading (1-3 Hz) for that "lost signal" feel.
	fadeStep := (1.0 + 2.0*norm) * 2 * math.Pi / fs
	for i, s := range samples {
		wide := float64(s)
		// Cascade two first-order high-pass sections (2nd order total).
		hp1 := aHP * (r.hp1 + wide - r.prevIn)
		hp2 := aHP * (r.hp2 + hp1 - r.hp1)
		r.prevIn = wide
		r.hp1 = hp1
		r.hp2 = hp2
		// Cascade two first-order low-pass sections (2nd order total).
		lp1 := r.lp1 + aLP*(hp2-r.lp1)
		lp2 := r.lp2 + aLP*(lp1-r.lp2)
		r.lp1 = lp1
		r.lp2 = lp2
		// Slow amplitude modulation (signal fading).
		band := lp2 * (0.85 + 0.15*math.Sin(r.fadePhase))
		// Breakup the band-limited signal: gain then fold back with tanh so
		// the voice gets the harsh mid-range "radio" edge.
		band = math.Tanh(band * drive)
		// Radio static grows with strength.
		band += r.nextNoise() * noiseGain
		samples[i] = float32(wide*(1-mix) + band*mix)
		r.fadePhase += fadeStep
		if r.fadePhase > 2*math.Pi {
			r.fadePhase -= 2 * math.Pi
		}
	}
}
