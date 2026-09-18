package gui

import (
	"fyne-io/fyne/v2/theme"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

type ThemeManager struct {
	currentTheme string
	onChange     func(themeName string)
}

func NewThemeManager() *ThemeManager {
	return &ThemeManager{
		currentTheme: "dark",
	}
}

func (t *ThemeManager) SetTheme(name string) {
	t.currentTheme = name
	if t.onChange != nil {
		t.onChange(name)
	}
}

func (t *ThemeManager) GetTheme() string {
	return t.currentTheme
}

func (t *ThemeManager) ApplyTo(app interface {
	Settings() theme.Settings
}) {
	if t.currentTheme == "light" {
		_ = app.Settings().SetTheme(theme.LightTheme())
	} else {
		_ = app.Settings().SetTheme(theme.DarkTheme())
	}
}

func (t *ThemeManager) IsDark() bool {
	return t.currentTheme == "dark"
}
