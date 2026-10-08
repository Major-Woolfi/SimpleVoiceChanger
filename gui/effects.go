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
		
		nameLabel := widget.NewLabel(eff.Name)
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
			catTitle := widget.NewLabel(e.app.t("category_"+cat))
			catTitle.TextStyle = fyne.TextStyle{Bold: true}
			sections = append(sections, catTitle)
			sections = append(sections, container.NewVBox(items...))
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
	descriptions := map[string]string{
		"antinoise":       "Адаптивное подавление шума на основе анализа спектра",
		"noise_suppressor": "Шумоподавитель с порогом (gate) для тихих участков",
		"low_cut":         "Высокочастотный фильтр (удаляет низкие частоты < 80Hz)",
		"high_cut":        "Низкочастотный фильтр (удаляет высокие частоты > 16kHz)",
		"formant":         "Сдвиг формант для изменения тембра голоса",
		"compressor":      "Динамический компрессор (ratio 4:1, attack/release)",
		"radio":           "Эмуляция радиосигнала с ограничением и шумом",
		"reverb":          "Реверберация с 4 задержками и обратной связью",
		"distortion":      "Мягкое насыщение/дисторшн через tanh",
		"saturation":      "Аналоговая насыщенность с экспоненциальной кривой",
		"harmony":         "Гармонизатор (добавляет октавы и квинты)",
		"chorus":          "Хорус с 3 линиями задержки",
		"flanger":         "Фланжер с модулируемой задержкой",
		"phaser":          "Фазер с 4 ступенями all-pass фильтров",
		"doubler":         "Даблер с легким детеуном и задержкой",
	}
	if _, ok := descriptions[key]; ok {
		return e.app.t(key + "_desc")
	}
	return ""
}

func (e *EffectsPanel) updateSliderLabel(l *widget.Label, eff effects.EffectInfo, value float64) {
	l.SetText(e.app.t("effect_strength_format", eff.Name, int(value+0.5)))
}