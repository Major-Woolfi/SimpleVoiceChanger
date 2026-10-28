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
	engine.SetOutputDeviceName(outputDevice)
	engine.SetOutput(playback)

	engine.RegisterNoiseEffect(effects.NewAntiNoise(sampleRate))
	engine.RegisterNoiseEffect(effects.NewNoiseSuppressor(sampleRate))
	engine.RegisterEffect(effects.NewLowCut(sampleRate))
	engine.RegisterEffect(effects.NewHighCut(sampleRate))
	engine.RegisterEffect(effects.NewFormant(sampleRate))
	engine.RegisterEffect(effects.NewCompressor(sampleRate))
	engine.RegisterEffect(effects.NewAmplifier())
	engine.RegisterEffect(effects.NewRadioEffect(sampleRate))
	engine.RegisterEffect(effects.NewReverb(sampleRate))
	engine.RegisterEffect(effects.NewDistortion(sampleRate))
	engine.RegisterEffect(effects.NewSaturation(sampleRate))
	engine.RegisterEffect(effects.NewHarmonyEngine(sampleRate))
	engine.RegisterEffect(effects.NewChorus(sampleRate))
	engine.RegisterEffect(effects.NewFlanger(sampleRate))
	engine.RegisterEffect(effects.NewPhaser(sampleRate))
	engine.RegisterEffect(effects.NewDoubler(sampleRate))

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
	engine.SetPitch(cfg.Pitch)

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
		newEngine.SetPitch(cfg.Pitch)
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
