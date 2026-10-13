package effects

import "math"

// Doubler thickens the voice with a short (5-10ms) modulated delay plus a
// gentle detune wobble, sounding like a double-tracked take.
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
	mix := 0.6 * norm
	lfoHz := 0.2 + 0.3*norm
	step := lfoHz * 2 * math.Pi / 48000
	n := d.bufSize
	for i, s := range samples {
		d.delayBuffer[d.bufPos] = s
		d.bufPos = (d.bufPos + 1) % n
		delay := 240.0 + 240.0*(0.5+0.5*math.Sin(d.phase))
		r := float64(d.bufPos) - delay
		i0 := int(math.Floor(r))
		frac := r - float64(i0)
		i0 = ((i0 % n) + n) % n
		i1 := (i0 + 1) % n
		doubled := float64(d.delayBuffer[i0])*(1-frac) + float64(d.delayBuffer[i1])*frac
		detune := float64(s) * (0.99 + 0.02*math.Sin(d.phase*5.0))
		samples[i] = float32(float64(s)*(1-mix) + mix*(doubled*0.6+detune*0.4))
		d.phase += step
		if d.phase > math.Pi*2 {
			d.phase -= math.Pi * 2
		}
	}
}
