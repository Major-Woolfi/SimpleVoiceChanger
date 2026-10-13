package effects

import "math"

type Formant struct {
	BaseEffect
	y1 float64
	y2 float64
	x1 float64
	x2 float64
}

func NewFormant() *Formant {
	return &Formant{
		BaseEffect: BaseEffect{
			name:     "formant",
			strength: 0,
			category: "modulation",
		},
	}
}

func (f *Formant) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.85 * norm
	// A resonant band-pass ("talk-box") whose centre shifts upward with
	// strength. Unity gain at the centre keeps the voice loud; the formant
	// colour changes instead of the voice fading away.
	centerHz := 500.0 + 900.0*norm
	resonance := 0.94 + 0.05*norm
	w := 2 * math.Pi * centerHz / 48000.0
	b1 := 2 * resonance * math.Cos(w)
	b2 := -resonance * resonance
	k := 1.0 - resonance*resonance
	for i, s := range samples {
		x := float64(s)
		y := k*(x-f.x2) + b1*f.y1 + b2*f.y2
		f.x2 = f.x1
		f.x1 = x
		f.y2 = f.y1
		f.y1 = y
		samples[i] = float32(x*(1.0-mix) + y*mix)
	}
}
