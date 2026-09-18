package effects

type EffectRegistry struct {
	Effects []EffectInfo
}

type EffectInfo struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

func DefaultEffects() []EffectInfo {
	return []EffectInfo{
		{Key: "antinoise", Name: "AntiNoise", Category: "noise"},
		{Key: "noise_suppressor", Name: "Noise Suppressor", Category: "noise"},
		{Key: "low_cut", Name: "Low Cut", Category: "filter"},
		{Key: "high_cut", Name: "High Cut", Category: "filter"},
		{Key: "formant", Name: "Formant", Category: "modulation"},
		{Key: "compressor", Name: "Compressor", Category: "dynamics"},
		{Key: "radio", Name: "Radio Effect", Category: "modulation"},
		{Key: "reverb", Name: "Reverb", Category: "spatial"},
		{Key: "distortion", Name: "Distortion", Category: "modulation"},
		{Key: "saturation", Name: "Saturation", Category: "modulation"},
		{Key: "harmony", Name: "Harmony Engine", Category: "voices"},
		{Key: "chorus", Name: "Chorus", Category: "voices"},
		{Key: "flanger", Name: "Flanger", Category: "voices"},
		{Key: "phaser", Name: "Phaser", Category: "voices"},
		{Key: "doubler", Name: "Doubler", Category: "voices"},
	}
}
