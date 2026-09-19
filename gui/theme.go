package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
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
	Settings() fyne.Settings
}) {
	if t.currentTheme == "light" {
		app.Settings().SetTheme(theme.LightTheme())
	} else {
		app.Settings().SetTheme(theme.DarkTheme())
	}
}

func (t *ThemeManager) IsDark() bool {
	return t.currentTheme == "dark"
}
