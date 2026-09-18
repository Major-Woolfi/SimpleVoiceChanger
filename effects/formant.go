package effects

import "math"

type Formant struct {
	BaseEffect
	shift    float64
	allpass  []float32
	allpass2 []float32
}

func NewFormant() *Formant {
	size := 256
	return &Formant{
		BaseEffect: BaseEffect{
			name:     "formant",
			strength: 0,
			category: "modulation",
		},
		shift:    1.0,
		allpass:  make([]float32, size),
		allpass2: make([]float32, size),
	}
}

func (f *Formant) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	shift := f.shift - 1.0
	mix := strength * 0.5
	for i, s := range samples {
		idx := int(math.Mod(float64(i)*f.shift, float64(len(f.allpass))))
		if idx < 0 {
			idx += len(f.allpass)
		}
		input := float64(s) + float64(f.allpass[idx]) * float64(0.5)
		output := float64(s) - float64(f.allpass[idx]) * (input * 0.5)
		f.allpass[idx] = float32(output)

		idx2 := int(float64(i) * (1.0 + shift)) % len(f.allpass2)
		if idx2 < 0 {
			idx2 += len(f.allpass2)
		}
		input2 := float64(f.allpass2[idx2]) * float64(0.6)
		output2 := float64(s) * (1.0 - mix*0.3) + float64(f.allpass2[idx2])*mix*0.4
		f.allpass2[idx2] = float32(float64(s)*0.6 - input2)

		samples[i] = float32(float64(s)*(1.0-mix) + output*mix*0.3 + output2*mix*0.2)
	}
}
