package audio

const (
	DefaultSampleRate = 48000
	DefaultBufferSize = 512
	DefaultChannels   = 1
)

type AudioBuffer struct {
	Data     []float32
	Channels int
	Rate     int
}

type EffectStrength float64

const (
	StrengthMin EffectStrength = 0.0
	StrengthMax EffectStrength = 1.0
)

func (s EffectStrength) Normalized() float64 {
	return float64(s) / 100.0
}

type ProcessFunc func(samples []float32, strength float64)

type EffectConfig struct {
	Name     string
	Strength EffectStrength
}
