package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AppConfig struct {
	Enabled       bool              `json:"enabled"`
	SampleRate    int               `json:"sample_rate"`
	BufferSize    int               `json:"buffer_size"`
	Theme         string            `json:"theme"`
	EffectStren   map[string]float64 `json:"effect_strengths"`
	SelectedPreset string           `json:"selected_preset"`
	ActivePresets []string          `json:"active_presets"`
	PresetsDir    string            `json:"-"`
	ConfigDir     string            `json:"-"`
}

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Enabled:     false,
		SampleRate:  48000,
		BufferSize:  512,
		Theme:       "dark",
		EffectStren: map[string]float64{},
		SelectedPreset: "",
		ActivePresets: []string{},
	}
}

func (c *AppConfig) Init(directories map[string]string) {
	c.PresetsDir = directories["presets"]
	c.ConfigDir = directories["config"]
	if c.EffectStren == nil {
		c.EffectStren = make(map[string]float64)
	}
	c.ensureDefaultStrengths()
}

func (c *AppConfig) ensureDefaultStrengths() {
	defaults := map[string]float64{
		"antinoise":        0,
		"noise_suppressor": 0,
		"low_cut":          0,
		"high_cut":         0,
		"formant":          0,
		"compressor":       0,
		"radio":            0,
		"reverb":           0,
		"distortion":       0,
		"saturation":       0,
		"harmony":          0,
		"chorus":           0,
		"flanger":          0,
		"phaser":           0,
		"doubler":          0,
	}
	for k, v := range defaults {
		if _, exists := c.EffectStren[k]; !exists {
			c.EffectStren[k] = v
		}
	}
}

func (c *AppConfig) Save() error {
	if c.ConfigDir == "" {
		return nil
	}
	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.ConfigDir, "config.json"), data, 0644)
}

func (c *AppConfig) Load() error {
	if c.ConfigDir == "" {
		return nil
	}
	path := filepath.Join(c.ConfigDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	c.ensureDefaultStrengths()
	return nil
}
