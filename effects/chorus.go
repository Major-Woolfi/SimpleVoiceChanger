package effects

import "math"

type Chorus struct {
	BaseEffect
	phase float64
	buf   []float32
	pos   int
}

func NewChorus() *Chorus {
	return &Chorus{
		BaseEffect: BaseEffect{
			name:     "chorus",
			strength: 0,
			category: "voices",
		},
		buf: make([]float32, 4800),
	}
}

func (c *Chorus) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.5 * norm
	// LFO sweeps the delay line between ~12ms and ~29ms at 0.25-0.7Hz so the
	// voice thickens smoothly instead of warbling (a fast LFO sounds broken).
	lfoHz := 0.25 + 0.45*norm
	step := lfoHz * 2 * math.Pi / 48000
	const base = 600
	const maxDepth = 800.0
	for i, s := range samples {
		c.buf[c.pos] = s
		c.pos = (c.pos + 1) % len(c.buf)
		delay := base + int(maxDepth*(0.5+0.5*math.Sin(c.phase)))
		read := c.pos - delay
		if read < 0 {
			read += len(c.buf)
		}
		wet := c.buf[read]
		samples[i] = float32(float64(s)*(1.0-mix) + float64(wet)*mix)
		c.phase += step
		if c.phase > 2*math.Pi {
			c.phase -= 2 * math.Pi
		}
	}
}
