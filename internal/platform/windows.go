package platform

type WindowsAudio struct{}

func (w *WindowsAudio) Init() error {
	return nil
}

func (w *WindowsAudio) Terminate() {
}
