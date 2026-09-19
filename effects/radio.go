package effects

type RadioEffect struct {
	BaseEffect
	state float32
}

func NewRadioEffect() *RadioEffect {
	return &RadioEffect{
		BaseEffect: BaseEffect{
			name:     "radio",
			strength: 0,
			category: "modulation",
		},
	}
}

func (r *RadioEffect) Process(samples []float32, strength float64) {
	if strength <= 0 {
		return
	}
	for i, s := range samples {
		clipped := float64(s) * (1.0 + strength*2.0)
		if clipped > 1.0 {
			clipped = 1.0
		}
		if clipped < -1.0 {
			clipped = -1.0
		}
		mixer := 0.7 * strength
		r.state += float32(float64(r.state) * 0.95)
		samples[i] = float32(clipped*mixer + float64(s)*(1.0-mixer) + float64(r.state)*0.05*strength)
	}
}
