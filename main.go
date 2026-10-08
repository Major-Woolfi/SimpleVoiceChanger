package main

import (
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

func main() {
	runtime.LockOSThread()

	app := gui.NewApp()

	cfg := core.DefaultAppConfig()
	cfg.Init(core.DefaultDirectories())
	_ = cfg.Load()

	engine := audio.NewEngine(cfg.SampleRate, cfg.BufferSize)

	capture := audio.NewCapture(cfg.SampleRate, cfg.BufferSize)
	if capture == nil {
		app.ShowError(app.T()("error"), app.T()("audio_init_failed")+": no capture backend")
		capture = audio.NewDefaultCapture()
	}
	engine.SetInput(capture)

	playback := audio.NewPlayback(cfg.SampleRate, cfg.BufferSize)
	if playback == nil {
		app.ShowError(app.T()("error"), app.T()("audio_init_failed")+": no playback backend")
		playback = audio.NewDefaultPlayback()
	}
	engine.SetOutput(playback)

	engine.RegisterNoiseEffect(effects.NewAntiNoise(cfg.SampleRate))
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

	for k, v := range cfg.EffectStren {
		engine.SetEffectStrength(k, v)
	}

	svc := service.NewService(engine)
	app.SetService(svc)
	app.SetConfig(cfg)

	app.SetOnToggle(func(enabled bool) {
		cfg.Enabled = enabled
		if enabled {
			for k, v := range cfg.EffectStren {
				engine.SetEffectStrength(k, v)
			}
		}
		_ = cfg.Save()
	})

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		app.Stop()
	}()

	app.Run()
}