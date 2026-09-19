package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

type PresetPanel struct {
	app    *App
	config *core.AppConfig
}

func NewPresetPanel(app *App, config *core.AppConfig) *PresetPanel {
	return &PresetPanel{
		app:    app,
		config: config,
	}
}

func (p *PresetPanel) Build() fyne.CanvasObject {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Имя пресета")

	saveBtn := widget.NewButton("Сохранить пресет", func() {
		if nameEntry.Text == "" {
			dialog.ShowInformation("", "Введите имя пресета", p.app.window)
			return
		}
		p.config.SelectedPreset = nameEntry.Text
		nameEntry.SetText("")
	})

	loadBtn := widget.NewButton("Загрузить пресет", func() {})

	listWidget := widget.NewList(
		func() int { return len(p.config.ActivePresets) },
		func() fyne.CanvasObject {
			return widget.NewLabel("пресет")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(p.config.ActivePresets[i])
		},
	)

	return container.NewBorder(
		widget.NewLabel("Управление пресетами"),
		container.NewVBox(layout.NewSpacer(), nameEntry, container.NewGridWithColumns(2, saveBtn, loadBtn)),
		nil, nil,
		listWidget,
	)
}
