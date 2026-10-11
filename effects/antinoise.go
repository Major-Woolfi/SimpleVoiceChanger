package effects

import "math"

type AntiNoise struct {
	BaseEffect
	sampleRate int
	floor      float32 // tracked noise floor of the room/mic
	gateState  float32
	openUntil  int
	floorDown  float64
	floorUp    float64
	gateAttack float64
	gateRel    float64
}

func NewAntiNoise(sampleRate int) *AntiNoise {
	return &AntiNoise{
		BaseEffect: BaseEffect{
			name:     "antinoise",
			strength: 0,
			category: "noise",
		},
		sampleRate: sampleRate,
		floor:      0.003,
		floorDown:  0.1,   // fast: a quieter frame lowers the estimate quickly
		floorUp:    0.002, // slow: drift the estimate back up between sentences
		gateAttack: 1.0 - math.Exp(-1.0/(float64(sampleRate)*0.005)),
		gateRel:    1.0 - math.Exp(-1.0/(float64(sampleRate)*0.15)),
	}
}

func (a *AntiNoise) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	idx := 0
	for idx < len(samples) {
		n := len(samples) - idx
		if n > 512 {
			n = 512
		}
		var sum float64
		for j := idx; j < idx+n; j++ {
			sum += float64(samples[j])
		}
		mean := sum / float64(n)
		var variance float64
		for j := idx; j < idx+n; j++ {
			d := float64(samples[j]) - mean
			variance += d * d
		}
		rms := math.Sqrt(variance / float64(n))
		floor := float64(a.floor)
		if rms < floor {
			floor += (rms - floor) * a.floorDown
		} else {
			floor += (rms - floor) * a.floorUp
		}
		a.floor = float32(floor)
		// The gate opens when a frame is clearly louder than the estimated
		// noise floor. Strength scales how aggressively quiet frames are cut.
		threshold := math.Max(0.006, floor*(1.4+1.6*norm))

		for j := idx; j < idx+n; j++ {
			amp := math.Abs(float64(samples[j]))
			if amp >= threshold {
				a.gateState = float32(float64(a.gateState) + (1.0-float64(a.gateState))*a.gateAttack)
				a.openUntil = j + a.sampleRate/8
			} else if j >= a.openUntil {
				a.gateState = float32(float64(a.gateState) + (0.0-float64(a.gateState))*a.gateRel)
			}
			if amp < threshold {
				samples[j] = float32(float64(samples[j]) * (0.1 + 0.15*norm))
			} else {
				samples[j] = float32(float64(samples[j]) * (1.0 - 0.05*norm))
			}
		}
		idx += n
	}
}
