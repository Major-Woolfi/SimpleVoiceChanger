package effects

import "math"

type NoiseSuppressor struct {
	BaseEffect
	envelope float64
	gate     float64
	attack   float64
	release  float64
}

func NewNoiseSuppressor() *NoiseSuppressor {
	return &NoiseSuppressor{
		BaseEffect: BaseEffect{
			name:     "noise_suppressor",
			strength: 0,
			category: "noise",
		},
		attack:  1.0 - math.Exp(-1.0/(48000.0*0.002)),
		release: 1.0 - math.Exp(-1.0/(48000.0*0.1)),
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
	threshold := 0.002 + 0.06*norm
	for i, s := range samples {
		level := math.Abs(float64(s))
		if level > n.envelope {
			n.envelope += (level - n.envelope) * n.attack
		} else {
			n.envelope += (level - n.envelope) * n.release
		}
		target := 1.0
		if n.envelope < threshold {
			target = 0.03 + 0.15*norm
		}
		if target > n.gate {
			n.gate += (target - n.gate) * n.attack
		} else {
			n.gate += (target - n.gate) * n.release
		}
		samples[i] = float32(float64(s) * n.gate)
	}
}
