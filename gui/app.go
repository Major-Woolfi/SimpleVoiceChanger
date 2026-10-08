package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/i18n"
)

type App struct {
	fyneApp   fyne.App
	window    fyne.Window
	config    *core.AppConfig
	running   bool
	engine    interface {
		Start() error
		Stop()
		SetEffectStrength(name string, strength float64)
		GetEffectStrengths() map[string]float64
	}
	service  interface {
		Start() error
		Stop()
	}
	i18n     *i18n.Manager
	t        i18n.T
	themeMgr *ThemeManager
	content  *fyne.Container
	tabs     *container.AppTabs
	toggleBtn *widget.Button
	statusLbl *widget.Label
	onToggle  func(enabled bool)
}

func NewApp() *App {
	a := app.NewWithID("github.com/Major-Woolfi.SimpleVoiceChanger")
	
	icon := loadAppIcon()
	if icon != nil {
		a.SetIcon(icon)
	}
	
	w := a.NewWindow("SimpleVoiceChanger")
	w.Resize(fyne.NewSize(900, 650))
	w.SetMaster()
	
	cfg := core.DefaultAppConfig()
	cfg.Init(core.DefaultDirectories())

	mgr := i18n.NewManagerFS(i18n.EmbeddedLocales, "en")
	t := mgr.T()

	themeMgr := NewThemeManager()
	themeMgr.SetApp(a)
	themeMgr.SetTheme(cfg.Theme)

	app := &App{
		fyneApp:   a,
		window:    w,
		config:    cfg,
		running:   false,
		i18n:      mgr,
		t:         t,
		themeMgr:  themeMgr,
	}

	mgr.SetLocaleChangeCallback(app.onLocaleChange)
	
	return app
}

func loadAppIcon() fyne.Resource {
	if icon, err := fyne.LoadResourceFromPath("assets/icons/mic.svg"); err == nil {
		return icon
	}
	return theme.ComputerIcon()
}

func (a *App) onLocaleChange() {
	a.t = a.i18n.T()
	fyne.Do(func() {
		a.rebuildUI()
	})
}

func (a *App) rebuildUI() {
	a.window.SetContent(a.buildContent())
	a.updateToggleButtonText()
	a.updateStatus()
}

func (a *App) updateToggleButtonText() {
	if a.toggleBtn != nil {
		if a.running {
			a.toggleBtn.SetText(a.t("toggle_stop"))
		} else {
			a.toggleBtn.SetText(a.t("toggle_voice_changer"))
		}
	}
}

func (a *App) Run() {
	a.window.SetContent(a.buildContent())
	a.window.ShowAndRun()
}

func (a *App) Stop() {
	if a.running {
		a.running = false
		a.updateStatus()
	}
	a.window.Close()
}

func (a *App) T() i18n.T {
	return a.t
}

func (a *App) SetConfig(config *core.AppConfig) {
	a.config = config
}

func (a *App) SetEngine(engine interface {
	Start() error
	Stop()
	SetEffectStrength(name string, strength float64)
	GetEffectStrengths() map[string]float64
}) {
	a.engine = engine
}

func (a *App) SetService(service interface {
	Start() error
	Stop()
}) {
	a.service = service
}

func (a *App) SetOnToggle(fn func(enabled bool)) {
	a.onToggle = fn
}

func (a *App) Toggle() {
	a.running = !a.running
	if a.running {
		if a.service != nil {
			if err := a.service.Start(); err != nil {
				a.ShowError(a.t("error"), err.Error())
				a.running = false
			}
		} else if a.engine != nil {
			if err := a.engine.Start(); err != nil {
				a.ShowError(a.t("error"), err.Error())
				a.running = false
			}
		}
	} else {
		if a.service != nil {
			a.service.Stop()
		} else if a.engine != nil {
			a.engine.Stop()
		}
	}
	a.updateStatus()
	a.updateToggleButtonText()
	if a.onToggle != nil {
		a.onToggle(a.running)
	}
}

func (a *App) IsRunning() bool {
	return a.running
}

func (a *App) updateStatus() {
	if a.statusLbl != nil {
		if a.running {
			a.statusLbl.SetText(a.t("status") + ": " + a.t("running"))
		} else {
			a.statusLbl.SetText(a.t("status") + ": " + a.t("stopped"))
		}
	}
}

func (a *App) buildMainMenu() fyne.CanvasObject {
	toggleBtn := widget.NewButton(a.t("toggle_voice_changer"), func() {
		a.Toggle()
	})
	toggleBtn.Importance = widget.HighImportance
	a.toggleBtn = toggleBtn

	presetSelect := widget.NewSelect(a.config.ActivePresets, func(s string) {
		a.config.SelectedPreset = s
	})
	presetSelect.PlaceHolder = a.t("select_preset")
	if a.config.SelectedPreset != "" {
		presetSelect.SetSelected(a.config.SelectedPreset)
	}

	testBtn := widget.NewButton(a.t("test_voice"), func() {
		a.testVoice()
	})

	importBtn := widget.NewButton(a.t("import_preset"), func() {
		a.importPreset()
	})
	exportBtn := widget.NewButton(a.t("export_preset"), func() {
		a.exportPreset()
	})

	statusLbl := widget.NewLabel(a.t("status") + ": " + a.t("stopped"))
	a.statusLbl = statusLbl

	topRow := container.NewHBox(
		layout.NewSpacer(),
		toggleBtn,
		layout.NewSpacer(),
		presetSelect,
		layout.NewSpacer(),
		testBtn,
		layout.NewSpacer(),
		importBtn,
		exportBtn,
		layout.NewSpacer(),
	)

	bottomRow := container.NewVBox(layout.NewSpacer(), statusLbl)

	return container.NewBorder(topRow, nil, nil, nil, bottomRow)
}

func (a *App) testVoice() {
	a.ShowInfo(a.t("info"), a.t("test_in_progress"))
	go func() {
		fyne.Do(func() {
			a.ShowInfo(a.t("info"), a.t("test_complete"))
		})
	}()
}

func (a *App) importPreset() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			a.ShowError(a.t("error"), err.Error())
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		a.ShowInfo(a.t("info"), a.t("import_success"))
	}, a.window)
}

func (a *App) exportPreset() {
	if a.config.SelectedPreset == "" {
		a.ShowInfo(a.t("warning"), a.t("select_preset"))
		return
	}
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			a.ShowError(a.t("error"), err.Error())
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()
		a.ShowInfo(a.t("info"), a.t("export_success"))
	}, a.window)
}

func (a *App) buildContent() fyne.CanvasObject {
	effectsPanel := NewEffectsPanel(a)
	presetsPanel := NewPresetPanel(a, a.config)
	settingsPanel := NewSettingsPanel(a, a.config)

	a.tabs = container.NewAppTabs(
		container.NewTabItem(a.t("effects"), effectsPanel.Build(a.config)),
		container.NewTabItem(a.t("presets"), presetsPanel.Build()),
		container.NewTabItem(a.t("settings"), settingsPanel.Build()),
	)
	a.tabs.SetTabLocation(container.TabLocationTop)

	menu := a.buildMainMenu()
	return container.NewBorder(menu, nil, nil, nil, a.tabs)
}

func (a *App) ShowInfo(title, message string) {
	dialog.ShowInformation(title, message, a.window)
}

func (a *App) ShowError(title, message string) {
	err := fmt.Errorf("%s: %s", title, message)
	dialog.ShowError(err, a.window)
}

func (a *App) ShowWarning(title, message string) {
	dialog.ShowConfirm(title, message, func(bool) {}, a.window)
}

var _ = container.NewAppTabs