package effects

import "math"

type NoiseSuppressor struct {
	BaseEffect
	sampleRate int
	envelope   float64
	gate       float64
	attack     float64
	release    float64
}

func NewNoiseSuppressor(sampleRate int) *NoiseSuppressor {
	return &NoiseSuppressor{
		BaseEffect: BaseEffect{
			name:     "noise_suppressor",
			strength: 0,
			category: "noise",
		},
		sampleRate: sampleRate,
		attack:     1.0 - math.Exp(-1.0/(float64(sampleRate)*0.002)),
		release:    1.0 - math.Exp(-1.0/(float64(sampleRate)*0.1)),
	}
}

// Process cuts quiet frames (room noise) toward the noise floor while keeping
// speech untouched. The gate follows a slow envelope so it never clicks or
// flaps on individual samples.
func (n *NoiseSuppressor) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	// Frames with an envelope below this absolute level are treated as noise.
	// Threshold scales gently so even at strength 100 normal quiet speech
	// is preserved; only near-silent room noise gets gated.
	threshold := 0.002 + 0.018*norm
	for i, s := range samples {
		level := math.Abs(float64(s))
		if level > n.envelope {
			n.envelope += (level - n.envelope) * n.attack
		} else {
			n.envelope += (level - n.envelope) * n.release
		}
		target := 1.0
		if n.envelope < threshold {
			target = 0.05 + 0.10*norm
		}
		if target > n.gate {
			n.gate += (target - n.gate) * n.attack
		} else {
			n.gate += (target - n.gate) * n.release
		}
		samples[i] = float32(float64(s) * n.gate)
	}
}
