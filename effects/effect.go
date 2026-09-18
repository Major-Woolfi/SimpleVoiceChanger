package effects

type Effect interface {
	Name() string
	Process(samples []float32, strength float64)
	SetStrength(s float64)
	Strength() float64
	Category() string
}

type BaseEffect struct {
	name     string
	strength float64
	category string
}

func (b *BaseEffect) Name() string {
	return b.name
}

func (b *BaseEffect) Strength() float64 {
	return b.strength
}

func (b *BaseEffect) SetStrength(s float64) {
	if s < 0 {
		s = 0
	}
	if s > 100 {
		s = 100
	}
	b.strength = s
}

func (b *BaseEffect) Category() string {
	return b.category
}
