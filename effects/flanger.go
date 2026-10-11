package effects

import "math"

type Flanger struct {
	BaseEffect
	phase   float64
	buffer  []float32
	bufSize int
	bufPos  int
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
	mix := 0.7 * norm
	// The delay sweeps from ~0.4ms to ~13ms at 0.1-0.4Hz (slow jet whoosh).
	lfoHz := 0.1 + 0.3*norm
	step := lfoHz * 2 * math.Pi / 48000
	for i, s := range samples {
		f.buffer[f.bufPos] = s
		f.bufPos = (f.bufPos + 1) % f.bufSize
		delaySamples := 20 + int(600.0*(0.5+0.5*math.Sin(f.phase)))
		if delaySamples >= f.bufSize {
			delaySamples = f.bufSize - 1
		}
		readPos := f.bufPos - delaySamples
		if readPos < 0 {
			readPos += f.bufSize
		}
		sampled := f.buffer[readPos]
		f.phase += step
		if f.phase > 2*math.Pi {
			f.phase -= 2 * math.Pi
		}
		samples[i] = float32(float64(s)*(1.0-mix) + float64(sampled)*mix)
	}
}
