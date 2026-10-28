package effects

import "math"

// Flanger sweeps a short (0.5-13ms) modulated delay with feedback. The deep
// mix and feedback produce the characteristic "jet" comb notches.
type Flanger struct {
	BaseEffect
	sampleRate int
	phase      float64
	buffer     []float32
	bufSize    int
	bufPos     int
	prev       float64
}

func NewFlanger(sampleRate int) *Flanger {
	s := float64(sampleRate) / 48000.0
	size := int(8192 * s)
	if size < 1 {
		size = 1
	}
	return &Flanger{
		BaseEffect: BaseEffect{
			name:     "flanger",
			strength: 0,
			category: "voices",
		},
		sampleRate: sampleRate,
		buffer:     make([]float32, size),
		bufSize:    size,
	}
}

func (f *Flanger) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.75 * norm
	feedback := 0.60 + 0.15*norm // 0.60→0.75 for deep comb resonance
	lfoHz := 0.1 + 0.7*norm
	step := lfoHz * 2 * math.Pi / float64(f.sampleRate)
	n := f.bufSize
	for i, s := range samples {
		f.buffer[f.bufPos] = float32(float64(s) + feedback*f.prev)
		f.bufPos = (f.bufPos + 1) % n
		// 20-300 samples (0.4-6 ms) for classic flanger comb notches.
		delay := 20.0 + 280.0*(0.5+0.5*math.Sin(f.phase))
		r := float64(f.bufPos) - delay
		i0 := int(math.Floor(r))
		frac := r - float64(i0)
		i0 = ((i0 % n) + n) % n
		i1 := (i0 + 1) % n
		wet := float64(f.buffer[i0])*(1-frac) + float64(f.buffer[i1])*frac
		// Normalize wet: comb with feedback peaks by (1+fb)/(1-fb);
		// multiply by (1-fb) to keep level constant.
		wet *= (1.0 - feedback)
		f.prev = wet
		samples[i] = float32(float64(s)*(1-mix) + wet*mix)
		f.phase += step
		if f.phase > 2*math.Pi {
			f.phase -= 2 * math.Pi
		}
	}
}
