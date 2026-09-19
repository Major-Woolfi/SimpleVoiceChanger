package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

type MainMenu struct {
	app        *App
	toggleBtn  *widget.Button
	presetBtn  *widget.Button
	testBtn    *widget.Button
	presetMenu *widget.PopUpMenu
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

	presetBtn := widget.NewButton("Выбрать пресет", func() {
		m.showPresetMenu(config)
	})
	presetBtn.Importance = widget.LowImportance

	topRow := container.New(
		layout.NewGridLayoutWithColumns(3),
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

func (m *MainMenu) showPresetMenu(config *core.AppConfig) {
	canvas := m.app.window.Canvas()
	if canvas == nil {
		return
	}

	items := []*fyne.MenuItem{}
	for _, name := range config.ActivePresets {
		name := name
		items = append(items, fyne.NewMenuItem(name, func() {
			config.SelectedPreset = name
		}))
	}
	if len(items) == 0 {
		items = append(items, fyne.NewMenuItem("Нет сохранённых пресетов", func() {}))
	}

	menu := widget.NewPopUpMenu(fyne.NewMenu("Пресеты", items...), canvas)
	menu.ShowAtPosition(fyne.NewPos(10, 10))
}
