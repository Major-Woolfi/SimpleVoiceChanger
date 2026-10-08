package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

type SettingsPanel struct {
	app             *App
	config          *core.AppConfig
	sampleRateEntry *widget.Entry
	bufferSizeEntry *widget.Entry
	themeSelect     *widget.Select
	langSelect      *widget.Select
}

func NewSettingsPanel(app *App, config *core.AppConfig) *SettingsPanel {
	return &SettingsPanel{app: app, config: config}
}

func (s *SettingsPanel) Build() fyne.CanvasObject {
	s.sampleRateEntry = widget.NewEntry()
	s.sampleRateEntry.SetText(strconv.Itoa(s.config.SampleRate))

	s.bufferSizeEntry = widget.NewEntry()
	s.bufferSizeEntry.SetText(strconv.Itoa(s.config.BufferSize))

	themeOptions := []string{s.app.t("dark"), s.app.t("light")}
	s.themeSelect = widget.NewSelect(themeOptions, func(theme string) {
		s.config.Theme = themeLabelToCode(theme)
		s.app.themeMgr.SetTheme(s.config.Theme)
	})
	s.themeSelect.SetSelected(themeCodeToLabel(s.config.Theme))

	// Language selector uses the raw locale codes.
	langCodes := []string{"en", "ru", "uk", "de", "fr", "es", "zh"}
	langLabels := make([]string, len(langCodes))
	for i, c := range langCodes {
		langLabels[i] = localeLabel(c)
	}
	s.langSelect = widget.NewSelect(langLabels, func(label string) {
		code := localeLabelToCode(label)
		s.app.i18n.SetLocale(code)
	})
	s.langSelect.SetSelected(localeLabel(s.app.i18n.Current()))

	inputBtn := widget.NewButton(s.app.t("input_device"), func() {})
	outputBtn := widget.NewButton(s.app.t("output_device"), func() {})

	saveBtn := widget.NewButton(s.app.t("ok"), func() {
		if v, err := strconv.Atoi(s.sampleRateEntry.Text); err == nil && v > 0 {
			s.config.SampleRate = v
		}
		if v, err := strconv.Atoi(s.bufferSizeEntry.Text); err == nil && v > 0 {
			s.config.BufferSize = v
		}
		_ = s.config.Save()
		s.app.ShowInfo(s.app.t("info"), s.app.t("config_saved"))
	})

	form := container.New(
		layout.NewFormLayout(),
		widget.NewLabel(s.app.t("sample_rate")),
		s.sampleRateEntry,
		widget.NewLabel(s.app.t("buffer_size")),
		s.bufferSizeEntry,
		widget.NewLabel(s.app.t("theme")),
		s.themeSelect,
		widget.NewLabel(s.app.t("language")),
		s.langSelect,
		widget.NewLabel(s.app.t("audio_device")),
		container.NewHBox(inputBtn, outputBtn),
	)

	return container.NewVBox(
		widget.NewLabel(s.app.t("settings")),
		form,
		container.NewHBox(layout.NewSpacer(), saveBtn, layout.NewSpacer()),
	)
}

func themeCodeToLabel(code string) string {
	if code == "light" {
		return "Light"
	}
	return "Dark"
}

func themeLabelToCode(label string) string {
	if label == "Light" {
		return "light"
	}
	return "dark"
}

func localeLabel(code string) string {
	switch code {
	case "ru":
		return "Русский"
	case "uk":
		return "Українська"
	case "de":
		return "Deutsch"
	case "fr":
		return "Français"
	case "es":
		return "Español"
	case "zh":
		return "中文"
	default:
		return "English"
	}
}

func localeLabelToCode(label string) string {
	switch label {
	case "Русский":
		return "ru"
	case "Українська":
		return "uk"
	case "Deutsch":
		return "de"
	case "Français":
		return "fr"
	case "Español":
		return "es"
	case "中文":
		return "zh"
	default:
		return "en"
	}
}
