package effects

import "math"

// HarmonyEngine weaves three short delay taps (13/17/21ms) with a slow
// 1-2 Hz detune wobble through interpolated reads, making the voice noticeably
// thicker and doubled without gliding to a different musical pitch.
type HarmonyEngine struct {
	BaseEffect
	sampleRate int
	phase      float64
	bufs       [][]float32
	pos        int
}

func NewHarmonyEngine(sampleRate int) *HarmonyEngine {
	s := float64(sampleRate) / 48000.0
	nBuf := int(4800 * s)
	if nBuf < 1 {
		nBuf = 1
	}
	b1 := make([]float32, nBuf)
	b2 := make([]float32, nBuf)
	b3 := make([]float32, nBuf)
	return &HarmonyEngine{
		BaseEffect: BaseEffect{
			name:     "harmony",
			strength: 0,
			category: "voices",
		},
		sampleRate: sampleRate,
		bufs:       [][]float32{b1, b2, b3},
	}
}

func (h *HarmonyEngine) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.7 * norm
	lfoHz := 1.0 + 1.0*norm
	step := lfoHz * 2 * math.Pi / float64(h.sampleRate)
	taps := 3
	delays := [3]float64{640, 810, 990}
	nBuf := len(h.bufs[0])
	for i, s := range samples {
		h.bufs[0][h.pos] = s
		h.bufs[1][h.pos] = s
		h.bufs[2][h.pos] = s
		h.pos = (h.pos + 1) % nBuf
		acc := float64(s) * (1.0 - mix)
		for j := 0; j < taps; j++ {
			wob := math.Sin(h.phase + float64(j)*2.1)
			delay := delays[j] * (1.0 + 0.06*wob)
			r := float64(h.pos) - delay
			i0 := int(math.Floor(r))
			frac := r - float64(i0)
			var i1 int
			if i0 < 0 {
				i0 += nBuf
			}
			i0 = i0 % nBuf
			i1 = (i0 + 1) % nBuf
			val := float64(h.bufs[j][i0])*(1-frac) + float64(h.bufs[j][i1])*frac
			acc += val * (1.0 + 0.008*wob) * mix / float64(taps)
		}
		h.phase += step
		if h.phase > 2*math.Pi {
			h.phase -= 2 * math.Pi
		}
		samples[i] = float32(acc)
	}
}
