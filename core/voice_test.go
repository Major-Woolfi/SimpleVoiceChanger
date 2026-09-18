package core

import (
	"fmt"
	"time"
)

type VoiceTestResult struct {
	Latency       time.Duration `json:"latency"`
	PeakLevel     float64       `json:"peak_level"`
	AvgLevel      float64       `json:"avg_level"`
	Clipping      bool          `json:"clipping"`
	Status        string        `json:"status"`
}

type VoiceTester struct {
	config *AppConfig
	engine interface {
		Start() error
		Stop()
		SetEffectStrength(name string, strength float64)
	}
}

func NewVoiceTester(config *AppConfig, engine interface {
	Start() error
	Stop()
	SetEffectStrength(name string, strength float64)
}) *VoiceTester {
	return &VoiceTester{
		config: config,
		engine: engine,
	}
}

func (vt *VoiceTester) RunTest(timeout time.Duration) (*VoiceTestResult, error) {
	if err := vt.engine.Start(); err != nil {
		return nil, fmt.Errorf("failed to start engine for test: %w", err)
	}
	defer vt.engine.Stop()

	result := &VoiceTestResult{
		Latency: timeout,
		Status:  "testing",
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}

	result.Status = "complete"
	result.PeakLevel = 0.0
	result.AvgLevel = 0.0
	result.Clipping = false

	return result, nil
}
