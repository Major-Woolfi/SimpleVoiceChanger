package gui

import (
	"fyne-io/fyne/v2"
	"fyne-io/fyne/v2/widget"
	"fyne-io/fyne/v2/container"
	"fyne-io/fyne/v2/layout"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

type EffectsPanel struct {
	app        *App
	sliders    map[string]*widget.Slider
	labels     map[string]*widget.Label
	effectsList []effects.EffectInfo
}

func NewEffectsPanel(app *App) *EffectsPanel {
	return &EffectsPanel{
		app:        app,
		sliders:    make(map[string]*widget.Slider),
		labels:     make(map[string]*widget.Label),
		effectsList: effects.DefaultEffects(),
	}
}

func (e *EffectsPanel) Build(config *core.AppConfig) fyne.CanvasObject {
	rows := make([]fyne.CanvasObject, 0, len(e.effectsList))
	for _, eff := range e.effectsList {
		slider := widget.NewSlider(0, 100)
		if val, ok := config.EffectStren[eff.Key]; ok {
			slider.SetValue(val)
		}
		slider.SetMinSize(fyne.NewSize(300, 30))
		slider.OnChange = func(value float64) {
			for k := range config.EffectStren {
				if k == eff.Key {
					config.EffectStren[k] = value
				}
			}
		}
		e.sliders[eff.Key] = slider
		e.labels[eff.Key] = widget.NewLabel(eff.Name)
		rows = append(rows, container.NewBorder(
			e.labels[eff.Key], nil, nil, layout.NewSpacer(),
			container.NewVBox(slider, widget.NewSeparator()),
		))
	}

	return container.NewVBox(
		widget.NewLabel("Эффекты"),
		container.New(layout.NewGridWithColumns(2), rows...),
	)
}
