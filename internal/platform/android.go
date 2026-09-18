package internal/platform

type AndroidAudio struct{}

func (a *AndroidAudio) Init() error {
	return nil
}

func (a *AndroidAudio) Terminate() {
}
