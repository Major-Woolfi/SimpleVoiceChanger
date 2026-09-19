package effects

import "math"

type AntiNoise struct {
	BaseEffect
	history    []float32
	sampleRate int
}

func NewAntiNoise(sampleRate int) *AntiNoise {
	return &AntiNoise{
		BaseEffect: BaseEffect{
			name:     "antinoise",
			strength: 0,
			category: "noise",
		},
		sampleRate: sampleRate,
		history:    make([]float32, 0, sampleRate/4),
	}
}

func (a *AntiNoise) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	windowSize := 256
	if len(samples) < windowSize {
		return
	}
	for i := 0; i <= len(samples)-windowSize; i += windowSize {
		end := i + windowSize
		if end > len(samples) {
			end = len(samples)
		}
		var sum float64
		for j := i; j < end; j++ {
			sum += float64(samples[j])
		}
		mean := sum / float64(end-i)
		var variance float64
		for j := i; j < end; j++ {
			diff := float64(samples[j]) - mean
			variance += diff * diff
		}
		variance /= float64(end - i)
		threshold := math.Sqrt(variance) * 1.5 * strength
		for j := i; j < end; j++ {
			amp := math.Abs(float64(samples[j]))
			if amp < threshold {
				reduction := amp / threshold
				samples[j] = float32(float64(samples[j]) * reduction * reduction * strength)
			} else {
				samples[j] = float32(float64(samples[j]) * (1.0 - (1.0-strength)*0.3))
			}
		}
	}
}
