package gui

import (
	"fyne-io/fyne/v2"
	"fyne-io/fyne/v2/widget"
	"fyne-io/fyne/v2/container"
	"fyne-io/fyne/v2/layout"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

type MainMenu struct {
	app        *App
	toggleBtn  *widget.Button
	presetBtn  *widget.Button
	testBtn    *widget.Button
	presetMenu *widget.PopupMenu
}

func NewMainMenu(app *App, config *core.AppConfig) *MainMenu {
	m := &MainMenu{
		app: app,
	}
	m.toggleBtn = widget.NewButton("", func() {
		app.Toggle()
	})
	m.presetBtn = widget.NewButton("", func() {})
	m.testBtn = widget.NewButton("", func() {})
	return m
}

func (m *MainMenu) Build(config *core.AppConfig) fyne.CanvasObject {
	m.toggleBtn.Text = "Включить голосовой чейнджер"
	m.presetBtn.Text = "Выбрать пресет"
	m.testBtn.Text = "Проверить голос"

	presets := []string{}
	for _, name := range config.ActivePresets {
		presets = append(presets, name)
	}
	if len(presets) == 0 {
		presets = append(presets, "Нет сохранённых пресетов")
	}
	m.presetMenu = widget.NewPopupMenu(presets, func(s string) {
		config.SelectedPreset = s
	})

	presetBtn := widget.NewButton("Выбрать пресет", func() {
		m.presetMenu.ShowAtPointer()
	})
	presetBtn.Importance = widget.LowImportance

	topRow := container.New(
		layout.NewGridWithColumns(3),
		container.NewVBox(m.toggleBtn),
		container.NewVBox(presetBtn),
		container.NewVBox(m.testBtn),
	)

	statusLabel := widget.NewLabel("Статус: Остановлен")
	if config.Enabled {
		statusLabel.SetText("Статус: Работает")
	}
	bottomRow := container.NewVBox(layout.NewSpacer(), statusLabel)

	return container.NewBorder(topRow, nil, nil, nil, bottomRow)
}
