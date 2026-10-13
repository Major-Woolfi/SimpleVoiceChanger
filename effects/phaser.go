package effects

import "math"

// Phaser is built from four cascaded one-pole allpass filters whose
// coefficients are swept by an LFO. The sweeping notch phase-cancellations
// give the classic phaser "whoosh". (The previous version was four plain
// smoothers, which simply dulled the voice.)
type Phaser struct {
	BaseEffect
	phase float64
	x     [4]float64
	y     [4]float64
}

func NewPhaser() *Phaser {
	return &Phaser{
		BaseEffect: BaseEffect{
			name:     "phaser",
			strength: 0,
			category: "voices",
		},
	}
}

func (p *Phaser) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.6 * norm
	lfoHz := 0.3 + 0.5*norm
	step := lfoHz * 2 * math.Pi / 48000
	for i, s := range samples {
		val := float64(s)
		for j := 0; j < 4; j++ {
			// Allpass coefficient sweeps 0.10..0.65 with per-stage phase offset.
			coef := 0.10 + 0.55*(0.5+0.5*math.Sin(p.phase+float64(j)*1.1))
			xo := val
			yo := coef*xo + p.x[j] - coef*p.y[j]
			p.x[j] = xo
			p.y[j] = yo
			val = yo
		}
		samples[i] = float32(float64(s)*(1-mix) + val*mix)
		p.phase += step
		if p.phase > 2*math.Pi {
			p.phase -= 2 * math.Pi
		}
	}
}
