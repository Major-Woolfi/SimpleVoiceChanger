package effects

import "math"

type Compressor struct {
	BaseEffect
	sampleRate int
	envelope   float64
}

func NewCompressor(sampleRate int) *Compressor {
	return &Compressor{
		BaseEffect: BaseEffect{
			name:     "compressor",
			strength: 0,
			category: "dynamics",
		},
		sampleRate: sampleRate,
		envelope:   1.0,
	}
}

func (c *Compressor) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	// Threshold drops as strength grows: more strength = compress more of the signal.
	threshold := 0.35 - 0.25*norm
	// Ratio grows with strength: 2:1 at low settings up to 10:1 at max.
	ratio := 2.0 + 8.0*norm
	attack := 1.0 - math.Exp(-1.0/(float64(c.sampleRate)*0.003))
	release := 1.0 - math.Exp(-1.0/(float64(c.sampleRate)*0.080))
	for i, s := range samples {
		// Envelope detector: smooth absolute value of input.
		absS := math.Abs(float64(s))
		var gain float64
		if absS > threshold {
			over := absS / threshold
			gain = math.Pow(over, 1.0/ratio-1.0)
		} else {
			gain = 1.0
		}
		// Smooth gain envelope: fast attack on peaks, slow release.
		if gain < c.envelope {
			c.envelope += (gain - c.envelope) * attack
		} else {
			c.envelope += (gain - c.envelope) * release
		}
		// Soft knee limiter instead of hard clip.
		out := float64(s) * c.envelope
		if out > 0.95 {
			out = 0.95 - (0.95-out)*(0.95-out)
		} else if out < -0.95 {
			out = -0.95 + (0.95+out)*(0.95+out)
		}
		samples[i] = float32(out)
	}
}
