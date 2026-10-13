package effects

import "math"

// Amplifier is a linear volume boost with a soft limiter so the signal never
// clips: strength 0 = unity (1x), strength 100 = up to 4x gain. The soft knee
// keeps mid-range amplification clean and only compresses peaks near the
// ceiling.
type Amplifier struct {
	BaseEffect
}

func NewAmplifier() *Amplifier {
	return &Amplifier{
		BaseEffect: BaseEffect{
			name:     "amplifier",
			strength: 0,
			category: "dynamics",
		},
	}
}

// Process applies gain = 1 + 3*strengthNorm and swallows peaks above the knee
// with an exponential soft limiter (asymptote at the ceiling).
func (g *Amplifier) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	gain := 1.0 + 3.0*norm
	const knee = 0.40
	const ceiling = 0.98
	headroom := ceiling - knee
	for i, x := range samples {
		s := float64(x) * gain
		a := math.Abs(s)
		if a > knee {
			excess := (a - knee) / headroom
			if excess > 0 {
				p := 1 - math.Exp(-excess)
				s = math.Copysign(knee+headroom*p, s)
			}
		}
		samples[i] = float32(s)
	}
}
