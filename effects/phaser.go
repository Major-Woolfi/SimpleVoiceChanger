package effects

import "math"

type Phaser struct {
	BaseEffect
	phase  float64
	stages []float32
}

func NewPhaser() *Phaser {
	return &Phaser{
		BaseEffect: BaseEffect{
			name:     "phaser",
			strength: 0,
			category: "voices",
		},
		stages: make([]float32, 4),
	}
}

func (p *Phaser) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.6 * norm
	// The allpass notches sweep at 0.3-0.8Hz for a slow, familiar phaser swell.
	lfoHz := 0.3 + 0.5*norm
	step := lfoHz * 2 * math.Pi / 48000
	for i, s := range samples {
		val := float64(s)
		for j := range p.stages {
			coef := 0.15 + 0.45*norm + math.Sin(p.phase+float64(j)*0.7)*0.03
			if coef < 0.02 {
				coef = 0.02
			}
			if coef > 0.9 {
				coef = 0.9
			}
			val = float64(p.stages[j]) + coef*(val-float64(p.stages[j]))
			p.stages[j] = float32(val)
		}
		p.phase += step
		if p.phase > 2*math.Pi {
			p.phase -= 2 * math.Pi
		}
		samples[i] = float32(float64(s)*(1.0-mix) + val*mix)
	}
}
