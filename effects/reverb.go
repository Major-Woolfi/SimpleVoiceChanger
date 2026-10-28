package effects

type Reverb struct {
	BaseEffect
	sampleRate int
	delays     [][]float32
	pos        []int
	lowpass    []float32
}

func NewReverb(sampleRate int) *Reverb {
	// Comb delay lengths scaled from a 48kHz reference to the actual rate so
	// the room size stays constant when the sample rate changes.
	s := float64(sampleRate) / 48000.0
	dl := []int{
		int(960 * s),
		int(1568 * s),
		int(2256 * s),
		int(2912 * s),
	}
	delays := make([][]float32, len(dl))
	for i, n := range dl {
		if n < 1 {
			dl[i] = 1
			n = 1
		}
		delays[i] = make([]float32, n)
	}
	return &Reverb{
		BaseEffect: BaseEffect{
			name:     "reverb",
			strength: 0,
			category: "spatial",
		},
		sampleRate: sampleRate,
		delays:     delays,
		pos:        make([]int, len(dl)),
		lowpass:    make([]float32, len(dl)),
	}
}

func (r *Reverb) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	norm := strength / 100
	mix := 0.45 * norm
	fb := 0.35 + 0.25*norm
	damp := 0.3 + 0.4*norm
	for i, s := range samples {
		acc := float64(s) * (1.0 - mix)
		for j := range r.delays {
			delayed := r.delays[j][r.pos[j]]
			// Damped feedback keeps the tail from getting metallic or unstable.
			lp := r.lowpass[j] + float32(damp)*(float32(float64(delayed)*fb)-r.lowpass[j])
			r.lowpass[j] = lp
			r.delays[j][r.pos[j]] = s + lp
			r.pos[j] = (r.pos[j] + 1) % len(r.delays[j])
			acc += float64(delayed) * mix / float64(len(r.delays))
		}
		samples[i] = float32(acc)
	}
}
