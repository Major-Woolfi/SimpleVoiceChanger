package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// T is a translator that resolves localized strings.
type T func(key string, args ...interface{}) string

// Locale holds loaded translations for a single language.
type Locale struct {
	Code    string
	Data    map[string]string
	Formats map[string]string
}

// Manager loads and caches locales.
type Manager struct {
	mu                sync.RWMutex
	locales           map[string]*Locale
	current           string
	fallback          string
	localeChangeCb    func()
}

// NewManager creates a manager and loads all JSON files from dir.
func NewManager(dir, fallback string) *Manager {
	m := &Manager{
		locales:  make(map[string]*Locale),
		fallback: fallback,
	}
	if err := m.LoadDir(dir); err == nil {
		m.current = fallback
	}
	return m
}

// NewManagerFS creates a manager and loads locales from an fs.FS
// (e.g. an embed.FS). This is the preferred constructor for installed
// builds where locale files are embedded in the binary.
func NewManagerFS(fsys fs.FS, fallback string) *Manager {
	m := &Manager{
		locales:  make(map[string]*Locale),
		fallback: fallback,
	}
	if err := m.LoadFS(fsys, "."); err == nil {
		m.current = fallback
	}
	return m
}

// LoadDir reads every *.json file in dir and registers it by file stem.
func (m *Manager) LoadDir(dir string) error {
	return m.LoadFS(os.DirFS(dir), ".")
}

// LoadFS reads every *.json file under dir from fsys.
func (m *Manager) LoadFS(fsys fs.FS, dir string) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		code := e.Name()[:len(e.Name())-len(filepath.Ext(e.Name()))]
		data, err := fs.ReadFile(fsys, filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		strs := make(map[string]string, len(raw))
		formats := make(map[string]string)
		for k, v := range raw {
			if s, ok := v.(string); ok {
				strs[k] = s
			}
		}
		for _, fk := range formatKeys {
			if v, ok := strs[fk]; ok {
				formats[fk] = v
				delete(strs, fk)
			}
		}
		m.locales[code] = &Locale{Code: code, Data: strs, Formats: formats}
	}
	return nil
}

// AvailableLocales returns the list of loaded locale codes.
func (m *Manager) AvailableLocales() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.locales))
	for k := range m.locales {
		out = append(out, k)
	}
	return out
}

// SetLocaleChangeCallback sets a callback to be called when locale changes.
func (m *Manager) SetLocaleChangeCallback(cb func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.localeChangeCb = cb
}

// SetLocale switches the active language.
func (m *Manager) SetLocale(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.locales[code]; ok {
		m.current = code
		if m.localeChangeCb != nil {
			go m.localeChangeCb()
		}
	}
}

// Current returns the active locale code.
func (m *Manager) Current() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// T returns a translator bound to the current locale.
func (m *Manager) T() T {
	return func(key string, args ...interface{}) string {
		m.mu.RLock()
		loc := m.locales[m.current]
		fallback := m.locales[m.fallback]
		m.mu.RUnlock()
		return translate(loc, fallback, key, args)
	}
}

func translate(loc, fallback *Locale, key string, args []interface{}) string {
	val := ""
	if loc != nil {
		val = loc.Data[key]
	}
	if val == "" && fallback != nil {
		val = fallback.Data[key]
	}
	if val == "" {
		return key
	}
	return interpolate(val, args)
}

func interpolate(format string, args []interface{}) string {
	if len(args) == 0 {
		return format
	}
	out := format
	for i, a := range args {
		ph := placeholderFor(i)
		out = replaceFirst(out, ph, argString(a))
	}
	return out
}

func placeholderFor(i int) string {
	ph := []string{"%s", "%d", "%v", "%g", "%x", "%o", "%t"}
	if i < len(ph) {
		return ph[i]
	}
	return "%s"
}

func argString(a interface{}) string {
	switch v := a.(type) {
	case string:
		return v
	case int:
		return itoa(v)
	case int64:
		return itoa(int(v))
	case float64:
		return ftoa(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", a)
	}
}

func replaceFirst(s, old, new string) string {
	idx := indexOf(s, old)
	if idx < 0 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	j := len(buf)
	for i > 0 {
		j--
		buf[j] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		j--
		buf[j] = '-'
	}
	return string(buf[j:])
}

func ftoa(f float64) string {
	return itoa(int(f))
}

var formatKeys = []string{
	"effect_strength_format",
}