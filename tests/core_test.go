package tests

import (
	"testing"

	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

func TestSanitizePresetName(t *testing.T) {
	valid := []string{"MyPreset", "Вокал 1", "preset_2024", "a-b.c (2)"}
	for _, name := range valid {
		if _, err := core.SanitizePresetName(name); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", name, err)
		}
	}

	invalid := []string{"", "  ", "..", "../evil", "a/b", "a\\b", "bad\x00", "x:", "name*star"}
	for _, name := range invalid {
		if _, err := core.SanitizePresetName(name); err == nil {
			t.Errorf("expected %q to be rejected", name)
		}
	}
}

func TestSavePresetRejectsTraversal(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultAppConfig()
	cfg.Init(map[string]string{"config": tmp})

	if err := cfg.SavePreset("../escape", map[string]float64{"reverb": 10}); err == nil {
		t.Fatal("expected traversal name to be rejected")
	}
	if err := cfg.SavePreset("safe", map[string]float64{"reverb": 10}); err != nil {
		t.Fatalf("safe preset should save: %v", err)
	}
}

// TestSyncPresetsDiscoversFiles verifies presets present on disk but missing
// from the config are surfaced after a restart.
func TestSyncPresetsDiscoversFiles(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultAppConfig()
	cfg.Init(map[string]string{"config": tmp})

	if err := cfg.SavePreset("DiskOnly", map[string]float64{"reverb": 5}); err != nil {
		t.Fatalf("save preset: %v", err)
	}
	if len(cfg.ActivePresets) != 0 {
		t.Fatalf("expected empty active presets, got %v", cfg.ActivePresets)
	}

	cfg.SyncPresets()
	found := false
	for _, n := range cfg.ActivePresets {
		if n == "DiskOnly" {
			found = true
		}
	}
	if !found {
		t.Fatalf("SyncPresets did not discover preset, got %v", cfg.ActivePresets)
	}

	cfg.SyncPresets()
	if len(cfg.ActivePresets) != 1 {
		t.Fatalf("SyncPresets should not duplicate entries, got %v", cfg.ActivePresets)
	}
}

// TestAppConfigDeviceRoundTrip verifies the selected input/output device names
// survive a save/load cycle (including non-ASCII names).
func TestAppConfigDeviceRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultAppConfig()
	cfg.Init(map[string]string{"config": tmp})
	cfg.InputDevice = "Микрофон (USB)"
	cfg.OutputDevice = "CABLE Input"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded := core.DefaultAppConfig()
	loaded.Init(map[string]string{"config": tmp})
	if err := loaded.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.InputDevice != "Микрофон (USB)" || loaded.OutputDevice != "CABLE Input" {
		t.Fatalf("device fields not persisted: in=%q out=%q", loaded.InputDevice, loaded.OutputDevice)
	}
}
