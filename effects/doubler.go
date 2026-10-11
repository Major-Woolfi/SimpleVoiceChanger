package effects

import "math"

type Doubler struct {
	BaseEffect
	phase       float64
	delayBuffer []float32
	bufSize     int
	bufPos      int
}

func NewDoubler() *Doubler {
	size := 2048
	return &Doubler{
		BaseEffect: BaseEffect{
			name:     "doubler",
			strength: 0,
			category: "voices",
		},
		delayBuffer: make([]float32, size),
		bufSize:     size,
	}
}

func (d *Doubler) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.45 * norm
	// The delay sits between ~5ms and ~10ms and drifts lazily at 0.2-0.5Hz,
	// which thickens the voice like an overdub rather than buzzing.
	lfoHz := 0.2 + 0.3*norm
	step := lfoHz * 2 * math.Pi / 48000
	for i, s := range samples {
		d.delayBuffer[d.bufPos] = s
		d.bufPos = (d.bufPos + 1) % d.bufSize
		delay := 240 + int(240.0*(0.5+0.5*math.Sin(d.phase)))
		readPos := d.bufPos - delay
		if readPos < 0 {
			readPos += d.bufSize
		}
		doubled := d.delayBuffer[readPos]
		detune := float64(s) * (0.99 + 0.01*math.Sin(d.phase*5.0))
		samples[i] = float32(float64(s)*(1.0-mix) + float64(doubled+float32(detune))*mix*0.5)
		d.phase += step
		if d.phase > math.Pi*2 {
			d.phase -= math.Pi * 2
		}
	}
}
