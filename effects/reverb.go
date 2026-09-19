package effects

type Reverb struct {
	BaseEffect
	delays   [][]float32
	delays2  [][]float32
	pos      []int
	pos2     []int
	feedback float64
}

func NewReverb() *Reverb {
	delays := []int{120, 240, 360, 480}
	d1 := make([]float32, delays[0])
	d2 := make([]float32, delays[1])
	d3 := make([]float32, delays[2])
	d4 := make([]float32, delays[3])
	return &Reverb{
		BaseEffect: BaseEffect{
			name:     "reverb",
			strength: 0,
			category: "spatial",
		},
		delays:   [][]float32{d1, d2, d3, d4},
		delays2:  [][]float32{d1, d2, d3, d4},
		pos:      []int{0, 0, 0, 0},
		pos2:     []int{0, 0, 0, 0},
		feedback: 0.5,
	}
}

func (r *Reverb) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	mix := strength * 0.5
	fb := r.feedback * strength
	for i, s := range samples {
		for d := 0; d < 4; d++ {
			delayed := r.delays[d][r.pos[d]]
			r.delays[d][r.pos[d]] = s + float32(float64(delayed)*fb)
			r.pos[d] = (r.pos[d] + 1) % len(r.delays[d])
			samples[i] += float32(float64(delayed) * mix * 0.25)
		}
	}
}
