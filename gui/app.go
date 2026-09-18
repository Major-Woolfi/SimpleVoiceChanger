package gui

import (
	"fyne-io/fyne/v2"
	"fyne-io/fyne/v2/app"
	"fyne-io/fyne/v2/theme"
	"fyne-io/fyne/v2/widget"
	"fyne-io/fyne/v2/container"
	"fyne-io/fyne/v2/layout"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window
	config  *core.AppConfig
	running bool
	engine interface {
		Start() error
		Stop()
		SetEffectStrength(name string, strength float64)
	}
	onToggle func(enabled bool)
}

func NewApp() *App {
	a := app.NewWithID("github.com.Major-Woolfi.SimpleVoiceChanger")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("SimpleVoiceChanger")
	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(true)
	return &App{
		fyneApp: a,
		window:  w,
		config:  core.DefaultAppConfig(),
		running: false,
	}
}

func (a *App) Run() {
	a.window.ShowAndRun()
}

func (a *App) Stop() {
	if a.running {
		a.running = false
	}
	a.window.Close()
}

func (a *App) SetConfig(config *core.AppConfig) {
	a.config = config
}

func (a *App) SetEngine(engine interface {
	Start() error
	Stop()
	SetEffectStrength(name string, strength float64)
}) {
	a.engine = engine
}

func (a *App) SetOnToggle(fn func(enabled bool)) {
	a.onToggle = fn
}

func (a *App) Toggle() {
	a.running = !a.running
	if a.onToggle != nil {
		a.onToggle(a.running)
	}
}

func (a *App) IsRunning() bool {
	return a.running
}

func (a *App) buildMainMenu() fyne.CanvasObject {
	toggleBtn := widget.NewButton("", func() {
		a.Toggle()
	})
	toggleBtn.Importance = widget.HighImportance

	presetSelect := widget.NewSelect([]string{}, func(s string) {})
	presetSelect.SetPlaceHolder("Выбрать пресет")

	testBtn := widget.NewButton("", func() {})

	importBtn := widget.NewButton("", func() {})
	exportBtn := widget.NewButton("", func() {})

	topRow := container.NewHBox(layout.NewSpacer(), toggleBtn, layout.NewSpacer(), presetSelect, layout.NewSpacer(), testBtn, layout.NewSpacer(), importBtn, exportBtn, layout.NewSpacer())
	_ = a.config
	return topRow
}

func (a *App) buildContent() fyne.CanvasObject {
	menu := a.buildMainMenu()
	return container.NewBorder(menu, nil, nil, nil, widget.NewLabel(""))
}
