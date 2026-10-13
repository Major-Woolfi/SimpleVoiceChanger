package effects

type Reverb struct {
	BaseEffect
	delays   [][]float32
	pos      []int
	lowpass  []float32
	feedback float64
}

func NewReverb() *Reverb {
	// Comb delays 20/32.7/47/60.7 ms give a room-like tail instead of a
	// metallic slap.
	dl := []int{960, 1568, 2256, 2912}
	delays := make([][]float32, len(dl))
	for i, n := range dl {
		delays[i] = make([]float32, n)
	}
	return &Reverb{
		BaseEffect: BaseEffect{
			name:     "reverb",
			strength: 0,
			category: "spatial",
		},
		delays:   delays,
		pos:      []int{0, 0, 0, 0},
		lowpass:  []float32{0, 0, 0, 0},
		feedback: 0.38,
	}
}

func (r *Reverb) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.45 * norm
	fb := 0.35 + 0.25*norm
	for i, s := range samples {
		acc := float64(s) * (1.0 - mix)
		for j := range r.delays {
			delayed := r.delays[j][r.pos[j]]
			// Damped feedback keeps the tail from getting metallic or unstable.
			lp := r.lowpass[j] + 0.3*(float32(float64(delayed)*fb)-r.lowpass[j])
			r.lowpass[j] = lp
			r.delays[j][r.pos[j]] = s + lp
			r.pos[j] = (r.pos[j] + 1) % len(r.delays[j])
			acc += float64(delayed) * mix / float64(len(r.delays))
		}
		samples[i] = float32(acc)
	}
}
