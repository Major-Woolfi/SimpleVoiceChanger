package effects

import "math"

type Formant struct {
	BaseEffect
	sampleRate int
	// Bandpass state for each formant frequency.
	y [][2]float64 // [formant][stage][y1, y2]
	x [][2]float64 // [formant][stage][x1, x2]
}

// formantFreqs are the four main vocal tract resonances in Hz.
var formantFreqs = []float64{500, 1220, 2300, 3400}

func NewFormant(sampleRate int) *Formant {
	n := len(formantFreqs)
	y := make([][2]float64, n)
	x := make([][2]float64, n)
	return &Formant{
		BaseEffect: BaseEffect{
			name:     "formant",
			strength: 0,
			category: "modulation",
		},
		sampleRate: sampleRate,
		y:          y,
		x:          x,
	}
}

// Process shifts vocal-formant resonances up or down, thickening or thinning
// the timbre. Multiple parallel resonators are summed to produce a natural
// "formant-shift" coloration rather than a single whistling peak.
func (f *Formant) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.85 * norm
	// Shift multiplier: 1.0 → 1.5 (lighter/younger voice) as strength grows.
	shift := 1.0 + 0.5*norm
	resonance := 0.72 + 0.18*norm
	n := len(formantFreqs)
	for i, s := range samples {
		x := float64(s)
		sum := x * (1.0 - mix)
		for j := 0; j < n; j++ {
			center := formantFreqs[j] * shift
			w := 2 * math.Pi * center / float64(f.sampleRate)
			b1 := 2 * resonance * math.Cos(w)
			b2 := -resonance * resonance
			k := 1.0 - resonance*resonance
			yj := k*(x-f.x[j][1]) + b1*f.y[j][0] + b2*f.y[j][1]
			f.y[j][1] = f.y[j][0]
			f.y[j][0] = yj
			f.x[j][1] = f.x[j][0]
			f.x[j][0] = x
			sum += float64(yj) * (1.0 / float64(n))
		}
		samples[i] = float32(sum)
	}
}
