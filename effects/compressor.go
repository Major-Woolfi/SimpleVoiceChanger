package effects

import "math"

type Compressor struct {
	BaseEffect
	envelope float64
}

func NewCompressor() *Compressor {
	return &Compressor{
		BaseEffect: BaseEffect{
			name:     "compressor",
			strength: 0,
			category: "dynamics",
		},
	}
}

func (c *Compressor) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	threshold := 0.25 + 0.05*norm
	ratio := 4.0
	// 5ms attack and 80ms release: fast enough to catch peaks, slow enough to
	// avoid the clicky "pumping" distortion that instant smoothing caused.
	attack := 1.0 - math.Exp(-1.0/(48000.0*0.005))
	release := 1.0 - math.Exp(-1.0/(48000.0*0.080))
	makeup := 1.0 + 0.3*norm
	for i, s := range samples {
		absS := math.Abs(float64(s))
		gain := 1.0
		if absS > threshold {
			over := absS / threshold
			gain = 1.0/over + (1.0-1.0/over)/ratio
		}
		if gain < c.envelope {
			c.envelope += (gain - c.envelope) * attack
		} else {
			c.envelope += (gain - c.envelope) * release
		}
		samples[i] = float32(float64(s) * (0.65 + 0.35*c.envelope) * makeup)
	}
}
