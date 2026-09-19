package effects

import "math"

type Chorus struct {
	BaseEffect
	phase float64
}

func NewChorus() *Chorus {
	return &Chorus{
		BaseEffect: BaseEffect{
			name:     "chorus",
			strength: 0,
			category: "voices",
		},
	}
}

func (c *Chorus) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	mix := strength * 0.5
	for i := range samples {
		for _, rate := range []float64{0.7, 0.9, 1.1} {
			offset := int(math.Sin(c.phase*rate)*8.0) + len(samples)/4
			if offset >= 0 && offset < len(samples) {
				weight := math.Sin(c.phase*rate) * mix * 0.3
				samples[i] += float32(float64(samples[offset]) * weight * strength)
			}
		}
		c.phase += 0.02 * strength
		if c.phase > math.Pi*2 {
			c.phase -= math.Pi * 2
		}
	}
}
