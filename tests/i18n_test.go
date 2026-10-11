package tests

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Major-Woolfi/SimpleVoiceChanger/i18n"
)

func TestI18nFormatKeyResolves(t *testing.T) {
	mgr := i18n.NewManagerFS(i18n.EmbeddedLocales, "en")
	tr := mgr.T()

	got := tr("effect_strength_format", "Reverb", 50)
	if strings.Contains(got, "effect_strength_format") {
		t.Fatalf("format key not resolved: %q", got)
	}
	if got != "Reverb: 50%" {
		t.Fatalf("unexpected format result: %q", got)
	}
}

func TestI18nFallbackForUnknownKey(t *testing.T) {
	mgr := i18n.NewManagerFS(i18n.EmbeddedLocales, "en")
	tr := mgr.T()
	if got := tr("this_key_does_not_exist"); got != "this_key_does_not_exist" {
		t.Fatalf("unknown key should return key itself, got %q", got)
	}
}

func TestI18nLocaleSwitch(t *testing.T) {
	mgr := i18n.NewManagerFS(i18n.EmbeddedLocales, "en")
	mgr.SetLocale("ru")
	if mgr.Current() != "ru" {
		t.Fatalf("locale not switched: %q", mgr.Current())
	}
	got := mgr.T()("effect_strength_format", "Эхо", 25)
	if got != "Эхо: 25%" {
		t.Fatalf("unexpected ru format: %q", got)
	}
}

// TestI18nSetLocaleOnlyFiresOnChange guards against the infinite UI-rebuild
// loop caused by Select.SetSelected firing OnChanged on every build.
func TestI18nSetLocaleOnlyFiresOnChange(t *testing.T) {
	mgr := i18n.NewManagerFS(i18n.EmbeddedLocales, "en")
	changes := make(chan string, 8)
	mgr.SetLocaleChangeCallback(func() {
		changes <- mgr.Current()
	})

	mgr.SetLocale("ru")
	mgr.SetLocale("ru")
	mgr.SetLocale("ru")
	select {
	case got := <-changes:
		if got != "ru" {
			t.Fatalf("unexpected first change: %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("expected a change callback for the first switch")
	}

	mgr.SetLocale("de")
	select {
	case got := <-changes:
		if got != "de" {
			t.Fatalf("unexpected second change: %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("expected a change callback for a real switch")
	}

	select {
	case extra := <-changes:
		t.Fatalf("unexpected extra callback for unchanged locale: %q", extra)
	case <-time.After(200 * time.Millisecond):
	}
}

// TestI18nLocaleKeyParity ensures every bundled locale exposes exactly the
// same set of keys as the English baseline, so no translation silently falls
// back for a missing key.
func TestI18nLocaleKeyParity(t *testing.T) {
	entries, err := fs.ReadDir(i18n.EmbeddedLocales, ".")
	if err != nil {
		t.Fatalf("read embedded locales: %v", err)
	}
	load := func(name string) map[string]string {
		raw, err := fs.ReadFile(i18n.EmbeddedLocales, name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		var m map[string]string
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("unmarshal %s: %v", name, err)
		}
		return m
	}

	base := load("en.json")
	if len(base) == 0 {
		t.Fatal("english baseline is empty")
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		m := load(e.Name())
		for k := range base {
			if _, ok := m[k]; !ok {
				t.Errorf("%s is missing key %q", e.Name(), k)
			}
		}
		for k := range m {
			if _, ok := base[k]; !ok {
				t.Errorf("%s has unexpected key %q", e.Name(), k)
			}
		}
	}
}
