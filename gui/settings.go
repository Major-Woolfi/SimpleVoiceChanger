package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/Major-Woolfi/SimpleVoiceChanger/audio"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/i18n"
)

type SettingsPanel struct {
	app             *App
	config          *core.AppConfig
	sampleRateEntry *widget.Entry
	bufferSizeEntry *widget.Entry
	inputSelect     *widget.Select
	outputSelect    *widget.Select
	themeSelect     *widget.Select
	langSelect      *widget.Select
	updating        bool
}

func NewSettingsPanel(app *App, config *core.AppConfig) *SettingsPanel {
	return &SettingsPanel{app: app, config: config}
}

func (s *SettingsPanel) Build() fyne.CanvasObject {
	defaultLabel := s.app.t("default_device")

	s.sampleRateEntry = widget.NewEntry()
	s.sampleRateEntry.SetText(strconv.Itoa(s.config.SampleRate))

	s.bufferSizeEntry = widget.NewEntry()
	s.bufferSizeEntry.SetText(strconv.Itoa(s.config.BufferSize))

	inputOptions := append([]string{defaultLabel}, audio.ListInputDevices()...)
	s.inputSelect = widget.NewSelect(inputOptions, func(label string) {
		if s.updating {
			return
		}
		s.config.InputDevice = deviceNameFromLabel(label, defaultLabel)
		_ = s.config.Save()
		s.app.rebuildAudio()
		s.app.RefreshDeviceStatus()
	})
	s.inputSelect.PlaceHolder = defaultLabel

	outputOptions := append([]string{defaultLabel}, audio.ListOutputDevices()...)
	s.outputSelect = widget.NewSelect(outputOptions, func(label string) {
		if s.updating {
			return
		}
		s.config.OutputDevice = deviceNameFromLabel(label, defaultLabel)
		_ = s.config.Save()
		s.app.rebuildAudio()
		s.app.RefreshDeviceStatus()
	})
	s.outputSelect.PlaceHolder = defaultLabel

	themeOptions := []string{s.app.t("light"), s.app.t("dark")}
	s.themeSelect = widget.NewSelect(themeOptions, func(theme string) {
		if s.updating {
			return
		}
		s.config.Theme = themeLabelToCode(s.app.t, theme)
		s.app.themeMgr.SetTheme(s.config.Theme)
		_ = s.config.Save()
	})

	// Language selector uses the raw locale codes.
	langCodes := []string{"en", "ru", "uk", "de", "fr", "es", "zh"}
	langLabels := make([]string, len(langCodes))
	for i, c := range langCodes {
		langLabels[i] = localeLabel(c)
	}
	s.langSelect = widget.NewSelect(langLabels, func(label string) {
		if s.updating {
			return
		}
		code := localeLabelToCode(label)
		s.config.Locale = code
		_ = s.config.Save()
		s.app.i18n.SetLocale(code)
	})

	// Programmatic SetSelected fires OnChanged in Fyne; suppress it while we
	// initialise the current values, otherwise the locale callback rebuilds
	// the whole UI and recurses forever.
	s.updating = true
	s.themeSelect.SetSelected(themeCodeToLabel(s.app.t, s.config.Theme))
	s.langSelect.SetSelected(localeLabel(s.app.i18n.Current()))
	s.selectInputDevice(defaultLabel)
	s.selectOutputDevice(defaultLabel)
	s.updating = false

	saveBtn := widget.NewButton(s.app.t("ok"), func() {
		oldSampleRate, oldBufferSize := s.config.SampleRate, s.config.BufferSize
		if v, err := strconv.Atoi(s.sampleRateEntry.Text); err == nil && v > 0 {
			s.config.SampleRate = v
		}
		if v, err := strconv.Atoi(s.bufferSizeEntry.Text); err == nil && v > 0 {
			s.config.BufferSize = v
		}
		_ = s.config.Save()
		if s.config.SampleRate != oldSampleRate || s.config.BufferSize != oldBufferSize {
			s.app.rebuildAudio()
		}
		s.app.ShowInfo(s.app.t("info"), s.app.t("config_saved"))
	})
	saveBtn.Importance = widget.HighImportance

	grid := container.New(
		newAdaptiveGridLayout(),
		s.card("input_device", s.inputSelect),
		s.card("output_device", s.outputSelect),
		s.card("sample_rate", s.sampleRateEntry),
		s.card("buffer_size", s.bufferSizeEntry),
		s.card("theme", s.themeSelect),
		s.card("language", s.langSelect),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle(s.app.t("settings"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		grid,
		container.NewHBox(layout.NewSpacer(), saveBtn),
	)
}

// card wraps a labelled control into a settings cell used by the adaptive grid.
func (s *SettingsPanel) card(labelKey string, control fyne.CanvasObject) fyne.CanvasObject {
	label := widget.NewLabel(s.app.t(labelKey))
	label.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewVBox(label, control)
}

func (s *SettingsPanel) selectInputDevice(defaultLabel string) {
	if s.config.InputDevice == "" {
		s.inputSelect.SetSelected(defaultLabel)
		return
	}
	s.inputSelect.SetSelected(s.config.InputDevice)
}

func (s *SettingsPanel) selectOutputDevice(defaultLabel string) {
	if s.config.OutputDevice == "" {
		s.outputSelect.SetSelected(defaultLabel)
		return
	}
	s.outputSelect.SetSelected(s.config.OutputDevice)
}

func deviceNameFromLabel(label, defaultLabel string) string {
	if label == defaultLabel {
		return ""
	}
	return label
}

func themeCodeToLabel(t i18n.T, code string) string {
	if code == "light" {
		return t("light")
	}
	return t("dark")
}

func themeLabelToCode(t i18n.T, label string) string {
	if label == t("light") {
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
