package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

type SettingsPanel struct {
	app    *App
	config *core.AppConfig
}

func NewSettingsPanel(app *App, config *core.AppConfig) *SettingsPanel {
	return &SettingsPanel{
		app:    app,
		config: config,
	}
}

func (s *SettingsPanel) Build() fyne.CanvasObject {
	sampleRateEntry := widget.NewEntry()
	sampleRateEntry.SetText("48000")

	bufferSizeEntry := widget.NewEntry()
	bufferSizeEntry.SetText("512")

	themeSelect := widget.NewSelect([]string{"dark", "light"}, func(theme string) {
		s.config.Theme = theme
	})
	themeSelect.SetSelected(s.config.Theme)

	autostartCheck := widget.NewCheck("", nil)
	minimizeTrayCheck := widget.NewCheck("", nil)

	return container.New(
		layout.NewFormLayout(),
		widget.NewLabel("Sample Rate"),
		sampleRateEntry,
		widget.NewLabel("Buffer Size"),
		bufferSizeEntry,
		widget.NewLabel("Theme"),
		themeSelect,
		widget.NewLabel("Autostart"),
		autostartCheck,
		widget.NewLabel("Minimize to tray"),
		minimizeTrayCheck,
	)
}
