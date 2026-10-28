package effects

import "math"

type AntiNoise struct {
	BaseEffect
	sampleRate int
	floor      float32 // tracked noise floor of the room/mic
	gateState  float32
	openUntil  int64 // absolute sample position the gate stays open until
	frames     int64 // absolute processed-sample counter
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
	hold := int64(a.sampleRate / 8)
	duck := 0.25 - 0.2*norm
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
		// noise floor. The gate gain is applied smoothly so the effect never
		// clicks, and the hold timer uses an absolute sample counter so it
		// survives buffer boundaries.
		threshold := math.Max(0.006, floor*(1.4+1.6*norm))

		for j := idx; j < idx+n; j++ {
			amp := math.Abs(float64(samples[j]))
			if amp >= threshold {
				a.openUntil = a.frames + hold
			}
			if a.frames <= a.openUntil {
				a.gateState = float32(float64(a.gateState) + (1.0-float64(a.gateState))*a.gateAttack)
			} else {
				a.gateState = float32(float64(a.gateState) + (0.0-float64(a.gateState))*a.gateRel)
			}
			gain := duck + (1.0-duck)*float64(a.gateState)
			samples[j] = float32(float64(samples[j]) * gain)
			a.frames++
		}
		idx += n
	}
}
