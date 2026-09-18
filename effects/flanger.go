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
	mix := strength * 0.6
	depth := 4.0 + strength*10.0
	rate := 0.1 + strength*0.4
	for i, s := range samples {
		f.buffer[f.bufPos] = s
		f.bufPos = (f.bufPos + 1) % f.bufSize
		delaySamples := int(depth + math.Sin(f.phase)*depth*0.5)
		if delaySamples < 0 {
			delaySamples = 0
		}
		if delaySamples >= f.bufSize {
			delaySamples = f.bufSize - 1
		}
		readPos := f.bufPos - delaySamples
		if readPos < 0 {
			readPos += f.bufSize
		}
		sampled := f.buffer[readPos]
		f.phase += rate * strength * 0.01
		if f.phase > math.Pi*2 {
			f.phase -= math.Pi * 2
		}
		samples[i] = float32(float64(s)*(1.0-mix*0.5) + float64(sampled)*mix*0.5)
	}
}
