package effects

import "math"

// Chorus is a classic modulated-delay chorus: three voices sweep through a
// shared delay line with fractional (interpolated) reads so the pitch bends
// smoothly instead of zipping. Mix grows with strength for a clearly audible
// thickening.
type Chorus struct {
	BaseEffect
	sampleRate int
	phase      float64
	buf        []float32
	pos        int
	prevWet    float64
}

func NewChorus(sampleRate int) *Chorus {
	s := float64(sampleRate) / 48000.0
	return &Chorus{
		BaseEffect: BaseEffect{
			name:     "chorus",
			strength: 0,
			category: "voices",
		},
		sampleRate: sampleRate,
		buf:        make([]float32, int(4800*s)),
	}
}

func (c *Chorus) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.55 * norm
	feedback := 0.35 * norm
	const base = 300.0  // ~6ms
	const depth = 700.0 // sweep to ~21ms
	lfoHz := 0.25 + 0.55*norm
	step := lfoHz * 2 * math.Pi / float64(c.sampleRate)
	n := len(c.buf)
	for i, s := range samples {
		c.buf[c.pos] = float32(float64(s) + feedback*c.prevWet)
		c.pos = (c.pos + 1) % n
		wet := 0.0
		for v := 0; v < 3; v++ {
			delay := base + depth*(0.5+0.5*math.Sin(c.phase+float64(v)*2.094))
			r := float64(c.pos) - delay
			i0 := int(math.Floor(r))
			frac := r - float64(i0)
			i0 = ((i0 % n) + n) % n
			i1 := (i0 + 1) % n
			wet += float64(c.buf[i0])*(1-frac) + float64(c.buf[i1])*frac
		}
		wet /= 3
		c.prevWet = wet
		samples[i] = float32(float64(s)*(1-mix) + wet*mix)
		c.phase += step
		if c.phase > 2*math.Pi {
			c.phase -= 2 * math.Pi
		}
	}
}
