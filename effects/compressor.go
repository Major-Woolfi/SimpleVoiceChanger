package effects

import "math"

type Compressor struct {
	BaseEffect
	ratio    float64
	attack   float64
	release  float64
	envelope float64
}

func NewCompressor() *Compressor {
	return &Compressor{
		BaseEffect: BaseEffect{
			name:     "compressor",
			strength: 0,
			category: "dynamics",
		},
		ratio:   4.0,
		attack:  0.005,
		release: 0.1,
	}
}

func (c *Compressor) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	threshold := 0.3 * strength
	for i, s := range samples {
		absS := math.Abs(float64(s))
		var gain float64 = 1.0
		if absS > threshold {
			excess := absS - threshold
			gain = 1.0 - (excess*(1.0-1.0/c.ratio)/absS)*strength
		}
		if gain > c.envelope {
			c.envelope += (gain - c.envelope) * c.attack
		} else {
			c.envelope += (gain - c.envelope) * c.release
		}
		samples[i] = float32(float64(s) * c.envelope * (0.5 + strength*0.5))
	}
}
