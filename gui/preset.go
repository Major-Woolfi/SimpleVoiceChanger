package gui

import (
	"encoding/json"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
)

type PresetPanel struct {
	app        *App
	config     *core.AppConfig
	nameEntry  *widget.Entry
	listWidget *widget.List
	selected   int
}

func NewPresetPanel(app *App, config *core.AppConfig) *PresetPanel {
	return &PresetPanel{app: app, config: config, selected: -1}
}

func (p *PresetPanel) Build() fyne.CanvasObject {
	p.nameEntry = widget.NewEntry()
	p.nameEntry.SetPlaceHolder(p.app.t("preset_name"))

	saveBtn := widget.NewButton(p.app.t("save_preset"), func() {
		p.savePreset()
	})

	loadBtn := widget.NewButton(p.app.t("load_preset"), func() {
		p.loadPreset()
	})

	deleteBtn := widget.NewButton(p.app.t("delete_preset"), func() {
		p.deletePreset()
	})

	importBtn := widget.NewButton(p.app.t("import_preset"), func() {
		p.importPreset()
	})

	exportBtn := widget.NewButton(p.app.t("export_preset"), func() {
		p.exportPreset()
	})

	p.listWidget = widget.NewList(
		func() int { return len(p.config.ActivePresets) },
		func() fyne.CanvasObject { return widget.NewLabel("preset") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(p.config.ActivePresets[i])
		},
	)
	p.listWidget.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(p.config.ActivePresets) {
			return
		}
		p.selected = id
		p.app.applyPresetSelection(p.config.ActivePresets[id])
	}

	return container.NewBorder(
		widget.NewLabel(p.app.t("presets")),
		container.NewVBox(
			layout.NewSpacer(),
			p.nameEntry,
			container.NewGridWithColumns(2, saveBtn, loadBtn),
			container.NewGridWithColumns(2, deleteBtn, importBtn),
			container.NewGridWithColumns(1, exportBtn),
		),
		nil, nil,
		p.listWidget,
	)
}

// SelectByName highlights the named preset in the list and fills the name
// entry, keeping the panel consistent with the main-menu selection.
func (p *PresetPanel) SelectByName(name string) {
	p.nameEntry.SetText(name)
	for i, n := range p.config.ActivePresets {
		if n == name {
			p.selected = i
			if p.listWidget != nil {
				p.listWidget.Select(i)
			}
			return
		}
	}
}

func (p *PresetPanel) savePreset() {
	name := p.nameEntry.Text
	if name == "" {
		p.app.ShowInfo(p.app.t("warning"), p.app.t("preset_name")+" "+p.app.t("error"))
		return
	}

	strengths := make(map[string]float64, len(effects.DefaultEffects()))
	for _, eff := range effects.DefaultEffects() {
		strengths[eff.Key] = p.config.EffectStren[eff.Key]
	}
	if p.app.engine != nil {
		for k, v := range p.app.engine.GetEffectStrengths() {
			strengths[k] = v
		}
	}

	err := p.config.SavePreset(name, strengths)
	if err != nil {
		p.app.ShowError(p.app.t("error"), err.Error())
		return
	}

	p.config.SelectedPreset = name
	p.config.ActivePresets = appendIfMissing(p.config.ActivePresets, name)
	p.nameEntry.SetText("")
	_ = p.config.Save()
	p.app.RefreshPresetSelect()
	p.app.ShowInfo(p.app.t("info"), p.app.t("preset_saved"))
	p.listWidget.Refresh()
}

func (p *PresetPanel) loadPreset() {
	if p.config.SelectedPreset == "" {
		p.app.ShowInfo(p.app.t("warning"), p.app.t("select_preset"))
		return
	}

	if err := p.app.loadPresetInto(p.config.SelectedPreset); err != nil {
		p.app.ShowError(p.app.t("error"), err.Error())
		return
	}

	p.app.ShowInfo(p.app.t("info"), p.app.t("preset_loaded"))
}

func (p *PresetPanel) deletePreset() {
	if p.config.SelectedPreset == "" {
		p.app.ShowInfo(p.app.t("warning"), p.app.t("select_preset"))
		return
	}

	name := p.config.SelectedPreset
	name, err := core.SanitizePresetName(name)
	if err != nil {
		p.app.ShowError(p.app.t("error"), err.Error())
		return
	}
	presetPath := filepath.Join(p.config.ConfigDir, "presets", name+".json")
	err = os.Remove(presetPath)
	if err != nil {
		p.app.ShowError(p.app.t("error"), err.Error())
		return
	}

	p.config.ActivePresets = removePreset(p.config.ActivePresets, name)
	p.config.SelectedPreset = ""
	p.nameEntry.SetText("")
	_ = p.config.Save()
	p.app.RefreshPresetSelect()
	p.app.ShowInfo(p.app.t("info"), p.app.t("preset_deleted"))
	p.listWidget.Refresh()
}

func (p *PresetPanel) importPreset() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			p.app.ShowError(p.app.t("error"), err.Error())
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		var preset core.EffectPreset
		decoder := json.NewDecoder(reader)
		if err := decoder.Decode(&preset); err != nil {
			p.app.ShowError(p.app.t("error"), err.Error())
			return
		}

		err = p.config.SavePresetWith(preset.Name, preset.Pitch, preset.Effects)
		if err != nil {
			p.app.ShowError(p.app.t("error"), err.Error())
			return
		}

		p.config.ActivePresets = appendIfMissing(p.config.ActivePresets, preset.Name)
		_ = p.config.Save()
		p.listWidget.Refresh()
		p.app.RefreshPresetSelect()
		p.app.ShowInfo(p.app.t("info"), p.app.t("import_success"))
	}, p.app.window)
}

func (p *PresetPanel) exportPreset() {
	if p.config.SelectedPreset == "" {
		p.app.ShowInfo(p.app.t("warning"), p.app.t("select_preset"))
		return
	}

	preset, err := p.config.LoadPreset(p.config.SelectedPreset)
	if err != nil {
		p.app.ShowError(p.app.t("error"), err.Error())
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			p.app.ShowError(p.app.t("error"), err.Error())
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(preset); err != nil {
			p.app.ShowError(p.app.t("error"), err.Error())
			return
		}
		p.app.ShowInfo(p.app.t("info"), p.app.t("export_success"))
	}, p.app.window)
}

func appendIfMissing(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

func removePreset(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
