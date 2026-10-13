package effects

import "math"

// Flanger sweeps a short (0.5-13ms) modulated delay with feedback. The deep
// mix and feedback produce the characteristic "jet" comb notches.
type Flanger struct {
	BaseEffect
	phase   float64
	buffer  []float32
	bufSize int
	bufPos  int
	prev    float64
}

func NewFlanger() *Flanger {
	size := 8192
	return &Flanger{
		BaseEffect: BaseEffect{
			name:     "flanger",
			strength: 0,
			category: "voices",
		},
		buffer:  make([]float32, size),
		bufSize: size,
	}
}

func (f *Flanger) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.75 * norm
	feedback := 0.30 * norm
	lfoHz := 0.1 + 0.35*norm
	step := lfoHz * 2 * math.Pi / 48000
	n := f.bufSize
	for i, s := range samples {
		f.buffer[f.bufPos] = float32(float64(s) + feedback*f.prev)
		f.bufPos = (f.bufPos + 1) % n
		delay := 24.0 + 640.0*(0.5+0.5*math.Sin(f.phase))
		r := float64(f.bufPos) - delay
		i0 := int(math.Floor(r))
		frac := r - float64(i0)
		i0 = ((i0 % n) + n) % n
		i1 := (i0 + 1) % n
		wet := float64(f.buffer[i0])*(1-frac) + float64(f.buffer[i1])*frac
		f.prev = wet
		samples[i] = float32(float64(s)*(1-mix) + wet*mix)
		f.phase += step
		if f.phase > 2*math.Pi {
			f.phase -= 2 * math.Pi
		}
	}
}
