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

	"github.com/Major-Woolfi/SimpleVoiceChanger/audio"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/i18n"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window
	config  *core.AppConfig
	running bool
	engine  interface {
		Start() error
		Stop()
		SetEffectStrength(name string, strength float64)
		GetEffectStrengths() map[string]float64
		SelfTest() error
		SetMonitor(on bool)
		SetPitch(semi float64)
	}
	service interface {
		Start() error
		Stop()
	}
	i18n              *i18n.Manager
	t                 i18n.T
	themeMgr          *ThemeManager
	content           *fyne.Container
	tabs              *container.AppTabs
	toggleBtn         *widget.Button
	statusLbl         *widget.Label
	presetPanel       *PresetPanel
	effectsPanel      *EffectsPanel
	presetSelect      *widget.Select
	testBtn           *widget.Button
	monitorCheckBox   *widget.Check
	inputStatusLbl    *widget.Label
	outputStatusLbl   *widget.Label
	testing           bool
	startedForTest    bool
	monitorBeforeTest bool // user's monitor setting before the test started
	syncing           bool
	onToggle          func(enabled bool)
	onSettingsApplied func()
}

func NewApp() *App {
	return newAppWith(app.NewWithID("com.majorwoolfi.simplevoicechanger"))
}

func newAppWith(a fyne.App) *App {
	icon := loadAppIcon()
	if icon != nil {
		a.SetIcon(icon)
	}

	w := a.NewWindow("Войсчейнжер")
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
		fyneApp:  a,
		window:   w,
		config:   cfg,
		running:  false,
		i18n:     mgr,
		t:        t,
		themeMgr: themeMgr,
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
	a.updateTestButtonText()
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

func (a *App) updateTestButtonText() {
	if a.testBtn != nil {
		if a.testing {
			a.testBtn.SetText(a.t("test_voice_stop"))
		} else {
			a.testBtn.SetText(a.t("test_voice"))
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

// SetLocale switches the UI language. It is a no-op for unknown codes.
func (a *App) SetLocale(code string) {
	if code != "" {
		a.i18n.SetLocale(code)
	}
}

func (a *App) SetConfig(config *core.AppConfig) {
	a.config = config
}

// loadPresetInto applies the named preset to the engine and the config,
// updates the effect sliders and persists the choice. Selecting a preset in
// the UI therefore takes effect immediately instead of only marking it.
func (a *App) loadPresetInto(name string) error {
	if name == "" {
		return nil
	}
	preset, err := a.config.LoadPreset(name)
	if err != nil {
		return err
	}
	if a.config.EffectStren == nil {
		a.config.EffectStren = make(map[string]float64)
	}
	for k, v := range preset.Effects {
		a.config.EffectStren[k] = v
		if a.engine != nil {
			a.engine.SetEffectStrength(k, v)
		}
	}
	a.config.Pitch = preset.Pitch
	if a.engine != nil {
		a.engine.SetPitch(preset.Pitch)
	}
	a.config.SelectedPreset = name
	_ = a.config.Save()
	a.RefreshEffectSliders()
	return nil
}

// applyPresetSelection keeps the main-menu select, the preset list and the
// name entry in sync when a preset is chosen anywhere in the UI, and applies
// its effects so the change is audible right away.
func (a *App) applyPresetSelection(name string) {
	if a.syncing {
		return
	}
	a.syncing = true
	defer func() { a.syncing = false }()

	a.config.SelectedPreset = name
	if a.presetSelect != nil {
		a.presetSelect.SetSelected(name)
	}
	if a.presetPanel != nil {
		a.presetPanel.SelectByName(name)
	}
	if err := a.loadPresetInto(name); err != nil {
		a.ShowError(a.t("error"), err.Error())
	}
}

// RefreshPresetSelect rebuilds the main-menu preset options after the
// preset list changes (save/delete/import) and restores the selection.
func (a *App) RefreshPresetSelect() {
	if a.presetSelect == nil {
		return
	}
	a.syncing = true
	defer func() { a.syncing = false }()

	a.presetSelect.Options = a.config.ActivePresets
	a.presetSelect.Refresh()
	if a.config.SelectedPreset != "" {
		a.presetSelect.SetSelected(a.config.SelectedPreset)
	} else {
		a.presetSelect.ClearSelected()
	}
}

// RefreshEffectSliders pushes the current config strengths into the effects
// tab sliders (used after loading a preset).
func (a *App) RefreshEffectSliders() {
	if a.effectsPanel != nil {
		a.effectsPanel.SyncFromConfig(a.config)
	}
}

func (a *App) SetEngine(engine interface {
	Start() error
	Stop()
	SetEffectStrength(name string, strength float64)
	GetEffectStrengths() map[string]float64
	SelfTest() error
	SetMonitor(on bool)
	SetPitch(semi float64)
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

// SetOnSettingsApplied registers a callback invoked when audio settings
// (sample rate, buffer size or devices) change so the audio chain can be
// rebuilt live.
func (a *App) SetOnSettingsApplied(fn func()) {
	a.onSettingsApplied = fn
}

// rebuildAudio asks the host to rebuild the audio chain from the current
// configuration (used after device or format changes).
func (a *App) rebuildAudio() {
	if a.onSettingsApplied != nil {
		a.onSettingsApplied()
	}
}

func (a *App) startAudio() error {
	if a.service != nil {
		return a.service.Start()
	}
	if a.engine != nil {
		return a.engine.Start()
	}
	return nil
}

func (a *App) stopAudio() {
	if a.service != nil {
		a.service.Stop()
	} else if a.engine != nil {
		a.engine.Stop()
	}
}

func (a *App) Toggle() {
	a.running = !a.running
	if a.running {
		if a.engine != nil {
			a.engine.SetMonitor(a.config.Monitor)
		}
		if err := a.startAudio(); err != nil {
			a.ShowError(a.t("error"), err.Error())
			a.running = false
		}
	} else {
		a.stopAudio()
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
	toggleBtn := widget.NewButtonWithIcon(a.t("toggle_voice_changer"), theme.MediaPlayIcon(), func() {
		a.Toggle()
	})
	toggleBtn.Importance = widget.HighImportance
	a.toggleBtn = toggleBtn

	presetSelect := widget.NewSelect(a.config.ActivePresets, func(s string) {
		a.applyPresetSelection(s)
	})
	presetSelect.PlaceHolder = a.t("select_preset")
	a.presetSelect = presetSelect
	if a.config.SelectedPreset != "" {
		a.syncing = true
		presetSelect.SetSelected(a.config.SelectedPreset)
		a.syncing = false
	}

	testBtn := widget.NewButtonWithIcon(a.t("test_voice"), theme.MediaRecordIcon(), func() {
		a.testVoice()
	})
	a.testBtn = testBtn

	monitorChk := widget.NewCheck(a.t("monitor"), func(on bool) {
		a.config.Monitor = on
		if a.engine != nil {
			a.engine.SetMonitor(on)
		}
		_ = a.config.Save()
	})
	monitorChk.SetChecked(a.config.Monitor)
	a.monitorCheckBox = monitorChk

	importBtn := widget.NewButtonWithIcon(a.t("import_preset"), theme.ContentAddIcon(), func() {
		a.importPreset()
	})
	exportBtn := widget.NewButtonWithIcon(a.t("export_preset"), theme.MailSendIcon(), func() {
		a.exportPreset()
	})

	statusLbl := widget.NewLabel(a.t("status") + ": " + a.t("stopped"))
	a.statusLbl = statusLbl

	inputStatusLbl := widget.NewLabel("")
	a.inputStatusLbl = inputStatusLbl
	outputStatusLbl := widget.NewLabel("")
	a.outputStatusLbl = outputStatusLbl

	a.RefreshDeviceStatus()

	// --- Top row: toggle, preset, test, monitor, import/export ---
	leftControls := container.NewHBox(
		toggleBtn,
		container.NewVBox(layout.NewSpacer(), presetSelect),
	)
	middleControls := container.NewHBox(testBtn, monitorChk)
	rightControls := container.NewHBox(importBtn, exportBtn)
	topBar := container.NewBorder(nil, nil, leftControls, rightControls, middleControls)

	// --- Status bar: status + device info ---
	deviceRow := container.NewHBox(
		container.NewHBox(widget.NewLabel("  "), inputStatusLbl),
		layout.NewSpacer(),
		container.NewHBox(widget.NewLabel("  "), outputStatusLbl),
	)
	statusRow := container.NewHBox(
		statusLbl,
		layout.NewSpacer(),
		deviceRow,
	)

	mainContent := container.NewBorder(topBar, nil, nil, nil, a.tabs)
	bottomBar := container.NewBorder(nil, statusRow, nil, nil, mainContent)
	return bottomBar
}

// testVoice toggles a live broadcast of the processed voice to the output
// device (speakers): the first press starts it, the second stops it.
func (a *App) testVoice() {
	if a.engine == nil {
		a.ShowError(a.t("error"), a.t("audio_init_failed"))
		return
	}
	if a.testing {
		// Restore the user's monitor preference before stopping the test.
		if a.engine != nil {
			a.engine.SetMonitor(a.monitorBeforeTest)
		}
		if a.startedForTest && !a.running {
			a.stopAudio()
		}
		a.testing = false
		a.updateTestButtonText()
		return
	}
	// Remember user's current monitor preference.
	a.monitorBeforeTest = a.config.Monitor
	a.startedForTest = !a.running
	if !a.running {
		if err := a.startAudio(); err != nil {
			a.ShowError(a.t("error"), err.Error())
			return
		}
	}
	if a.engine != nil {
		a.engine.SetMonitor(true)
	}
	a.testing = true
	a.updateTestButtonText()
}

func (a *App) importPreset() {
	if a.presetPanel == nil {
		a.ShowError(a.t("error"), "preset panel unavailable")
		return
	}
	a.presetPanel.importPreset()
}

func (a *App) exportPreset() {
	if a.presetPanel == nil {
		a.ShowError(a.t("error"), "preset panel unavailable")
		return
	}
	a.presetPanel.exportPreset()
}

// buildContent builds the tabbed UI. Panels are created once and cached in
// the App struct; on locale change rebuildUI updates the tab titles without
// recreating the widgets, so slider values, preset selection and form state
// survive a language switch.
func (a *App) buildContent() fyne.CanvasObject {
	if a.tabs != nil && a.effectsPanel != nil && a.presetPanel != nil {
		a.updateTabTitles()
		a.tabs.Refresh()
		return a.buildMainMenu()
	}

	effectsPanel := NewEffectsPanel(a)
	presetsPanel := NewPresetPanel(a, a.config)
	a.presetPanel = presetsPanel
	a.effectsPanel = effectsPanel
	settingsPanel := NewSettingsPanel(a, a.config)

	a.tabs = container.NewAppTabs(
		container.NewTabItem(a.t("effects"), effectsPanel.Build(a.config)),
		container.NewTabItem(a.t("presets"), presetsPanel.Build()),
		container.NewTabItem(a.t("settings"), settingsPanel.Build()),
	)
	a.tabs.SetTabLocation(container.TabLocationTop)

	return a.buildMainMenu()
}

// updateTabTitles refreshes the tab captions after a locale change.
func (a *App) updateTabTitles() {
	if a.tabs == nil || len(a.tabs.Items) < 3 {
		return
	}
	a.tabs.Items[0].Text = a.t("effects")
	a.tabs.Items[1].Text = a.t("presets")
	a.tabs.Items[2].Text = a.t("settings")
}

// RefreshDeviceStatus updates the input/output device status labels
// shown below the main menu bar. It also reports the built-in loopback capture
// device, which is what a chat program has to pick as its microphone to hear
// the processed voice.
func (a *App) RefreshDeviceStatus() {
	if a.inputStatusLbl == nil || a.outputStatusLbl == nil {
		return
	}
	inputDevices := audio.ListInputDevices()
	outputDevices := audio.ListOutputDevices()
	inputMsg := a.t("device_found")
	if len(inputDevices) == 0 {
		inputMsg = a.t("device_not_found")
	}
	outputMsg := a.t("device_found")
	if len(outputDevices) == 0 {
		outputMsg = a.t("device_not_found")
	}
	// The processed voice always goes to the selected output device. A chat
	// program reads it back through the built-in mix recorder (Stereo Mix),
	// so the status line names that recorder when the system provides one.
	outDev := a.config.OutputDevice
	if outDev == "" {
		outDev = a.t("default_device")
	}
	mix := audio.FindBuiltinLoopback()
	micHint := a.t("virtual_mic_missing")
	if mix != "" {
		micHint = a.t("virtual_mic_found", mix)
	}
	a.inputStatusLbl.SetText(a.t("input_status", inputMsg))
	a.outputStatusLbl.SetText(a.t("output_status", outputMsg) +
		" | " + a.t("output_device") + ": " + outDev + " | " + micHint)
}

func (a *App) ShowInfo(title, message string) {
	dialog.ShowInformation(title, message, a.window)
}

func (a *App) ShowError(title, message string) {
	err := fmt.Errorf("%s: %s", title, message)
	dialog.ShowError(err, a.window)
}
