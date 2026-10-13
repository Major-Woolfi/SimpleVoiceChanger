package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

type EffectsPanel struct {
	app         *App
	sliders     map[string]*widget.Slider
	labels      map[string]*widget.Label
	effectsList []effects.EffectInfo
}

func NewEffectsPanel(app *App) *EffectsPanel {
	return &EffectsPanel{
		app:         app,
		sliders:     make(map[string]*widget.Slider),
		labels:      make(map[string]*widget.Label),
		effectsList: effects.DefaultEffects(),
	}
}

func (e *EffectsPanel) Build(config *core.AppConfig) fyne.CanvasObject {
	categoryContainers := make(map[string][]fyne.CanvasObject)
	categoryOrder := []string{"noise", "filter", "modulation", "dynamics", "spatial", "voices"}

	for _, eff := range e.effectsList {
		slider := widget.NewSlider(0, 100)
		if val, ok := config.EffectStren[eff.Key]; ok {
			slider.SetValue(val)
		}
		sliderLabel := widget.NewLabel("")
		e.updateSliderLabel(sliderLabel, eff, slider.Value)
		slider.OnChanged = func(value float64) {
			config.EffectStren[eff.Key] = value
			e.updateSliderLabel(sliderLabel, eff, value)
			if e.app.engine != nil {
				e.app.engine.SetEffectStrength(eff.Key, value)
			}
		}
		e.sliders[eff.Key] = slider
		e.labels[eff.Key] = sliderLabel

		nameLabel := widget.NewLabel(e.effectName(eff))
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}

		descLabel := widget.NewLabel(e.getEffectDescription(eff.Key))
		descLabel.TextStyle = fyne.TextStyle{Italic: true}
		descLabel.Wrapping = fyne.TextWrapWord

		categoryBox := container.NewVBox(
			nameLabel,
			descLabel,
			slider,
			sliderLabel,
			widget.NewSeparator(),
		)
		categoryContainers[eff.Category] = append(categoryContainers[eff.Category], categoryBox)
	}

	sections := make([]fyne.CanvasObject, 0, len(categoryOrder))
	for _, cat := range categoryOrder {
		if items, ok := categoryContainers[cat]; ok && len(items) > 0 {
			catTitle := widget.NewLabel(e.app.t("category_" + cat))
			catTitle.TextStyle = fyne.TextStyle{Bold: true}
			sections = append(sections, catTitle)
			sections = append(sections, container.NewGridWithColumns(2, items...))
			sections = append(sections, widget.NewSeparator())
		}
	}

	return container.NewVScroll(
		container.NewVBox(
			append([]fyne.CanvasObject{widget.NewLabel(e.app.t("effects"))}, sections...)...,
		),
	)
}

func (e *EffectsPanel) getEffectDescription(key string) string {
	for _, eff := range e.effectsList {
		if eff.Key == key {
			return e.app.t(key + "_desc")
		}
	}
	return ""
}

// SyncFromConfig updates every slider and its label from the given config,
// so the effects tab reflects strengths loaded from a preset or on startup.
func (e *EffectsPanel) SyncFromConfig(config *core.AppConfig) {
	for _, eff := range e.effectsList {
		slider, ok := e.sliders[eff.Key]
		if !ok {
			continue
		}
		val := config.EffectStren[eff.Key]
		slider.SetValue(val)
		if lbl, ok := e.labels[eff.Key]; ok {
			e.updateSliderLabel(lbl, eff, val)
		}
		slider.Refresh()
	}
}

func (e *EffectsPanel) updateSliderLabel(l *widget.Label, eff effects.EffectInfo, value float64) {
	l.SetText(e.app.t("effect_strength_format", e.effectName(eff), int(value+0.5)))
}

// effectName returns the localized effect name, falling back to the
// registry name when no translation is available.
func (e *EffectsPanel) effectName(eff effects.EffectInfo) string {
	key := eff.Key + "_name"
	name := e.app.t(key)
	if name == key {
		return eff.Name
	}
	return name
}
