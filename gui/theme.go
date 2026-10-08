package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeManager tracks the active theme and applies it to the Fyne app.
type ThemeManager struct {
	currentTheme string
	onChange     func(themeName string)
	app          fyne.App
}

func NewThemeManager() *ThemeManager {
	return &ThemeManager{currentTheme: "dark"}
}

func (t *ThemeManager) SetApp(a fyne.App) {
	t.app = a
}

func (t *ThemeManager) SetTheme(name string) {
	t.currentTheme = name
	if t.app != nil {
		if name == "light" {
			t.app.Settings().SetTheme(theme.LightTheme())
		} else {
			t.app.Settings().SetTheme(theme.DarkTheme())
		}
	}
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
	return t.currentTheme != "light"
}
