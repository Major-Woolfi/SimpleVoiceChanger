package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unicode"
)

// SanitizePresetName validates a preset name and returns it trimmed or an
// error if it contains path separators, "..", control characters or other
// unsafe characters. Prevents path traversal when joining with the
// presets directory.
func SanitizePresetName(name string) (string, error) {
	trimmed := trimmedPresetName(name)
	if trimmed == "" || trimmed == "." || trimmed == ".." {
		return "", fmt.Errorf("invalid preset name")
	}
	if filepath.Base(trimmed) != trimmed {
		return "", fmt.Errorf("preset name must not contain path separators")
	}
	if containsUnsafe(trimmed) {
		return "", fmt.Errorf("preset name contains unsafe characters")
	}
	return trimmed, nil
}

func trimmedPresetName(name string) string {
	start := 0
	end := len(name)
	for start < end && (name[start] == ' ' || name[start] == '\t' || name[start] == '\r' || name[start] == '\n') {
		start++
	}
	for end > start && (name[end-1] == ' ' || name[end-1] == '\t' || name[end-1] == '\r' || name[end-1] == '\n') {
		end--
	}
	return name[start:end]
}

func containsUnsafe(s string) bool {
	for _, r := range s {
		if r == '/' || r == '\\' || r < 0x20 {
			return true
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
			r != ' ' && r != '_' && r != '-' && r != '.' && r != '(' && r != ')' {
			return true
		}
	}
	return false
}

// AppConfig holds the application configuration.
type AppConfig struct {
	ConfigDir      string
	SampleRate     int
	BufferSize     int
	InputDevice    string
	OutputDevice   string
	Theme          string
	Locale         string
	Enabled        bool
	Monitor        bool
	Pitch          float64
	EffectStren    map[string]float64
	SelectedPreset string
	ActivePresets  []string
}

// EffectPreset is a named collection of effect strengths.
type EffectPreset struct {
	Name    string             `json:"name"`
	Pitch   float64            `json:"pitch"`
	Effects map[string]float64 `json:"effects"`
}

// DefaultAppConfig returns a new AppConfig with sensible defaults.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		SampleRate:  48000,
		BufferSize:  512,
		Theme:       "dark",
		Enabled:     false,
		Monitor:     true,
		EffectStren: make(map[string]float64),
	}
}

// Init sets up directories and ensures they exist.
func (c *AppConfig) Init(dirs map[string]string) {
	if v, ok := dirs["config"]; ok {
		c.ConfigDir = v
	}
	if c.ConfigDir == "" {
		c.ConfigDir = "./SimpleVoiceChanger"
	}
	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		// best-effort; continue
	}
	if v, ok := dirs["presets"]; ok {
		if err := os.MkdirAll(v, 0755); err != nil {
			// best-effort
		}
	}
}

// configFileName returns the path to the JSON config file.
func (c *AppConfig) configFileName() string {
	return filepath.Join(c.ConfigDir, "config.json")
}

// Save persists the configuration to disk.
func (c *AppConfig) Save() error {
	data := map[string]interface{}{
		"enabled":          c.Enabled,
		"monitor":          c.Monitor,
		"pitch":            c.Pitch,
		"sample_rate":      c.SampleRate,
		"buffer_size":      c.BufferSize,
		"input_device":     c.InputDevice,
		"output_device":    c.OutputDevice,
		"theme":            c.Theme,
		"locale":           c.Locale,
		"effect_strengths": c.EffectStren,
		"selected_preset":  c.SelectedPreset,
		"active_presets":   c.ActivePresets,
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(c.configFileName(), raw, 0644)
}

// Load reads the configuration from disk.
func (c *AppConfig) Load() error {
	raw, err := os.ReadFile(c.configFileName())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config: %w", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	if v, ok := data["enabled"].(bool); ok {
		c.Enabled = v
	}
	if v, ok := data["monitor"].(bool); ok {
		c.Monitor = v
	}
	if v, ok := data["pitch"].(float64); ok {
		c.Pitch = v
	}
	if v, ok := data["sample_rate"].(float64); ok {
		c.SampleRate = int(v)
	}
	if v, ok := data["buffer_size"].(float64); ok {
		c.BufferSize = int(v)
	}
	if v, ok := data["input_device"].(string); ok {
		c.InputDevice = v
	}
	if v, ok := data["output_device"].(string); ok {
		c.OutputDevice = v
	}
	if v, ok := data["theme"].(string); ok {
		c.Theme = v
	}
	if v, ok := data["locale"].(string); ok {
		c.Locale = v
	}
	if v, ok := data["selected_preset"].(string); ok {
		c.SelectedPreset = v
	}
	if v, ok := data["active_presets"].([]interface{}); ok {
		c.ActivePresets = make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				c.ActivePresets = append(c.ActivePresets, s)
			}
		}
	}
	if v, ok := data["effect_strengths"].(map[string]interface{}); ok {
		if c.EffectStren == nil {
			c.EffectStren = make(map[string]float64)
		}
		for k, val := range v {
			if f, ok := val.(float64); ok {
				c.EffectStren[k] = f
			}
		}
	}
	return nil
}

// SavePreset persists a named preset to disk, capturing the current pitch.
func (c *AppConfig) SavePreset(name string, strengths map[string]float64) error {
	return c.SavePresetWith(name, c.Pitch, strengths)
}

// SavePresetWith persists a named preset with an explicit pitch offset. It is
// used by import, where the pitch must come from the file rather than the
// live configuration.
func (c *AppConfig) SavePresetWith(name string, pitch float64, strengths map[string]float64) error {
	name, err := SanitizePresetName(name)
	if err != nil {
		return err
	}
	presetsDir := filepath.Join(c.ConfigDir, "presets")
	if err := os.MkdirAll(presetsDir, 0755); err != nil {
		return fmt.Errorf("create presets dir: %w", err)
	}
	p := EffectPreset{Name: name, Pitch: pitch, Effects: strengths}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal preset: %w", err)
	}
	return os.WriteFile(filepath.Join(presetsDir, name+".json"), raw, 0644)
}

// LoadPreset reads a named preset from disk.
func (c *AppConfig) LoadPreset(name string) (*EffectPreset, error) {
	name, err := SanitizePresetName(name)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(c.ConfigDir, "presets", name+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read preset: %w", err)
	}
	var p EffectPreset
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("unmarshal preset: %w", err)
	}
	return &p, nil
}

// ListPresets returns the names of all saved presets.
func (c *AppConfig) ListPresets() ([]string, error) {
	dir := filepath.Join(c.ConfigDir, "presets")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list presets: %w", err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out = append(out, e.Name()[:len(e.Name())-len(filepath.Ext(e.Name()))])
	}
	return out, nil
}

// SyncPresets merges preset files found on disk into ActivePresets so that
// presets present in the presets directory are always visible after restart.
func (c *AppConfig) SyncPresets() {
	names, err := c.ListPresets()
	if err != nil {
		return
	}
	known := make(map[string]bool, len(c.ActivePresets))
	for _, n := range c.ActivePresets {
		known[n] = true
	}
	changed := false
	for _, n := range names {
		if n == "" || known[n] {
			continue
		}
		c.ActivePresets = append(c.ActivePresets, n)
		known[n] = true
		changed = true
	}
	if changed {
		_ = c.Save()
	}
}
