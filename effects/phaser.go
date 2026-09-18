package effects

import "math"

type Phaser struct {
	BaseEffect
	phase  float64
	stages []float32
	fpos   []int
}

func NewPhaser() *Phaser {
	stages := 4
	stg := make([]float32, stages)
	fp := make([]int, stages)
	for i := range fp {
		fp[i] = 0
	}
	return &Phaser{
		BaseEffect: BaseEffect{
			name:     "phaser",
			strength: 0,
			category: "voices",
		},
		stages: stg,
		fpos:   fp,
	}
}

func (p *Phaser) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	mix := strength * 0.6
	for i, s := range samples {
		val := float64(s)
		for j := 0; j < len(p.stages); j++ {
			f := 0.5 + math.Sin(p.phase+float64(j)*0.5)*0.3*strength
			alpha := float32(f * 0.1 * strength)
			val = float64(p.stages[j]) + alpha*(val - float64(p.stages[j]))
			p.stages[j] = float32(val)
			p.fpos[j] = (p.fpos[j] + 1) % 64
		}
		p.phase += 0.008 * strength
		if p.phase > math.Pi*2 {
			p.phase -= math.Pi * 2
		}
		samples[i] += float32(float64(s)*mix*0.3)
	}
}
