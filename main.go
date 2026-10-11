package main

import (
	"errors"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/Major-Woolfi/SimpleVoiceChanger/audio"
	"github.com/Major-Woolfi/SimpleVoiceChanger/core"
	"github.com/Major-Woolfi/SimpleVoiceChanger/effects"
	"github.com/Major-Woolfi/SimpleVoiceChanger/gui"
	"github.com/Major-Woolfi/SimpleVoiceChanger/internal/service"
)

// buildEngine creates a full audio pipeline (capture, playback and effects)
// for the given sample rate, buffer size and devices. It always returns a
// usable engine; the returned error reports a missing platform backend while
// the engine falls back to the built-in mock devices.
func buildEngine(sampleRate, bufferSize int, inputDevice, outputDevice string) (*audio.Engine, error) {
	engine := audio.NewEngine(sampleRate, bufferSize)
	var initErr error

	capture := audio.NewCapture(sampleRate, bufferSize, inputDevice)
	if capture == nil {
		initErr = errors.New("no capture backend")
		capture = audio.NewDefaultCapture()
	}
	engine.SetInput(capture)

	playback := audio.NewPlayback(sampleRate, bufferSize, outputDevice)
	if playback == nil {
		if initErr == nil {
			initErr = errors.New("no playback backend")
		}
		playback = audio.NewDefaultPlayback()
	}
	engine.SetOutput(playback)

	engine.RegisterNoiseEffect(effects.NewAntiNoise(sampleRate))
	engine.RegisterNoiseEffect(effects.NewNoiseSuppressor())
	engine.RegisterEffect(effects.NewLowCut())
	engine.RegisterEffect(effects.NewHighCut())
	engine.RegisterEffect(effects.NewFormant())
	engine.RegisterEffect(effects.NewCompressor())
	engine.RegisterEffect(effects.NewRadioEffect())
	engine.RegisterEffect(effects.NewReverb())
	engine.RegisterEffect(effects.NewDistortion())
	engine.RegisterEffect(effects.NewSaturation())
	engine.RegisterEffect(effects.NewHarmonyEngine())
	engine.RegisterEffect(effects.NewChorus())
	engine.RegisterEffect(effects.NewFlanger())
	engine.RegisterEffect(effects.NewPhaser())
	engine.RegisterEffect(effects.NewDoubler())

	return engine, initErr
}

func main() {
	runtime.LockOSThread()

	app := gui.NewApp()

	cfg := core.DefaultAppConfig()
	cfg.Init(core.DefaultDirectories())
	_ = cfg.Load()
	cfg.SyncPresets()

	engine, initErr := buildEngine(cfg.SampleRate, cfg.BufferSize, cfg.InputDevice, cfg.OutputDevice)
	if initErr != nil {
		app.ShowError(app.T()("error"), app.T()("audio_init_failed")+": "+initErr.Error())
	}
	for k, v := range cfg.EffectStren {
		engine.SetEffectStrength(k, v)
	}
	engine.SetMonitor(cfg.Monitor)

	svc := service.NewService(engine)
	app.SetService(svc)
	app.SetEngine(engine)
	app.SetConfig(cfg)

	locale := cfg.Locale
	if locale == "" {
		locale = detectSystemLocale()
	}
	if locale != "" {
		app.SetLocale(locale)
	}

	app.SetOnToggle(func(enabled bool) {
		cfg.Enabled = enabled
		if enabled {
			for k, v := range cfg.EffectStren {
				engine.SetEffectStrength(k, v)
			}
		}
		_ = cfg.Save()
	})

	app.SetOnSettingsApplied(func() {
		wasRunning := app.IsRunning()
		if wasRunning {
			svc.Stop()
		}
		newEngine, rebuildErr := buildEngine(cfg.SampleRate, cfg.BufferSize, cfg.InputDevice, cfg.OutputDevice)
		if rebuildErr != nil {
			app.ShowError(app.T()("error"), app.T()("audio_init_failed")+": "+rebuildErr.Error())
		}
		for k, v := range cfg.EffectStren {
			newEngine.SetEffectStrength(k, v)
		}
		newEngine.SetMonitor(cfg.Monitor)
		engine = newEngine
		svc = service.NewService(engine)
		app.SetEngine(engine)
		app.SetService(svc)
		if wasRunning {
			if err := svc.Start(); err != nil {
				app.ShowError(app.T()("error"), err.Error())
			}
		}
	})

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		app.Stop()
	}()

	app.Run()
}
