package internal.service

import (
	"sync"

	"github.com/Major-Woolfi/SimpleVoiceChanger/audio"
)

type Service struct {
	engine *audio.Engine
	mu     sync.RWMutex
	running bool
}

func NewService(engine *audio.Engine) *Service {
	return &Service{
		engine: engine,
	}
}

func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil
	}
	if err := s.engine.Start(); err != nil {
		return err
	}
	s.running = true
	return nil
}

func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.engine.Stop()
	s.running = false
}
