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
	mix := strength * 0.4
	for i, s := range samples {
		d.delayBuffer[d.bufPos] = s
		d.bufPos = (d.bufPos + 1) % d.bufSize
		delay := 10 + int(math.Sin(d.phase)*8)
		if delay < 0 {
			delay = 0
		}
		readPos := d.bufPos - delay
		if readPos < 0 {
			readPos += d.bufSize
		}
		if readPos >= d.bufSize {
			readPos = 0
		}
		doubled := d.delayBuffer[readPos]
		detune := float64(s) * (0.99 + math.Sin(d.phase*3.0)*0.01*strength)
		samples[i] = float32(float64(s)*(1.0-mix) + float64(doubled+float32(detune))*mix*0.5)
		d.phase += 0.003 * strength
		if d.phase > math.Pi*2 {
			d.phase -= math.Pi * 2
		}
	}
}
