package effects

import "math"

type HarmonyEngine struct {
	BaseEffect
	phase float64
	bufs  [][]float32
	pos   int
}

func NewHarmonyEngine() *HarmonyEngine {
	b1 := make([]float32, 4800)
	b2 := make([]float32, 4800)
	b3 := make([]float32, 4800)
	return &HarmonyEngine{
		BaseEffect: BaseEffect{
			name:     "harmony",
			strength: 0,
			category: "voices",
		},
		bufs: [][]float32{b1, b2, b3},
	}
}

func (h *HarmonyEngine) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.5 * norm
	// Three short taps (13/17/21ms) with a subtle 1-2Hz detune wobble weave the
	// voice into a thicker, doubled sound - no fast warble, no pitch jump.
	lfoHz := 1.0 + 1.0*norm
	step := lfoHz * 2 * math.Pi / 48000
	const taps = 3
	delays := [taps]int{640, 810, 990}
	lfoPhase := h.phase
	for i, s := range samples {
		h.bufs[0][h.pos] = s
		h.bufs[1][h.pos] = s
		h.bufs[2][h.pos] = s
		h.pos = (h.pos + 1) % len(h.bufs[0])
		acc := float64(s) * (1.0 - mix)
		for j := 0; j < taps; j++ {
			read := h.pos - delays[j]
			if read < 0 {
				read += len(h.bufs[j])
			}
			detsub := math.Sin(lfoPhase + float64(j)*2.1)
			ratio := 1.0 + 0.008*detsub
			acc += float64(h.bufs[j][read]) * ratio * mix / float64(taps)
		}
		h.phase += step
		if h.phase > 2*math.Pi {
			h.phase -= 2 * math.Pi
		}
		samples[i] = float32(acc)
	}
}
