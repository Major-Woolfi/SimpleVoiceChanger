package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type EffectPreset struct {
	Name    string             `json:"name"`
	Effects map[string]float64 `json:"effects"`
	Created string             `json:"created,omitempty"`
}

type PresetManager struct {
	presetsDir string
	presets    map[string]*EffectPreset
}

func NewPresetManager(presetsDir string) *PresetManager {
	return &PresetManager{
		presetsDir: presetsDir,
		presets:    make(map[string]*EffectPreset),
	}
}

func (pm *PresetManager) LoadAll() error {
	if pm.presetsDir == "" {
		return nil
	}
	entries, err := os.ReadDir(pm.presetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(pm.presetsDir, entry.Name()))
		if err != nil {
			continue
		}
		var preset EffectPreset
		if err := json.Unmarshal(data, &preset); err != nil {
			continue
		}
		pm.presets[preset.Name] = &preset
	}
	return nil
}

func (pm *PresetManager) SavePreset(preset *EffectPreset) error {
	if pm.presetsDir == "" {
		return nil
	}
	if err := os.MkdirAll(pm.presetsDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return err
	}
	name := preset.Name
	safeName := makeSafeFileName(name)
	return os.WriteFile(filepath.Join(pm.presetsDir, safeName+".json"), data, 0644)
}

func (pm *PresetManager) DeletePreset(name string) error {
	if pm.presetsDir == "" {
		return nil
	}
	path := filepath.Join(pm.presetsDir, makeSafeFileName(name)+".json")
	return os.Remove(path)
}

func (pm *PresetManager) ListPresets() []string {
	names := make([]string, 0, len(pm.presets))
	for name := range pm.presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (pm *PresetManager) GetPreset(name string) *EffectPreset {
	return pm.presets[name]
}

func (pm *PresetManager) ImportPreset(data []byte) (*EffectPreset, error) {
	var preset EffectPreset
	if err := json.Unmarshal(data, &preset); err != nil {
		return nil, err
	}
	pm.presets[preset.Name] = &preset
	return &preset, nil
}

func (pm *PresetManager) ExportPreset(name string) ([]byte, error) {
	preset := pm.presets[name]
	if preset == nil {
		return nil, os.ErrNotExist
	}
	return json.MarshalIndent(preset, "", "  ")
}

func makeSafeFileName(name string) string {
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result = append(result, r)
		} else if r == ' ' {
			result = append(result, '_')
		}
	}
	return string(result)
}
