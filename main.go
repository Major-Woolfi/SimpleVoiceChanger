package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Major-Woolfi/SimpleVoiceChanger/gui"
)

func main() {
	app := gui.NewApp()
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		app.Stop()
	}()
	app.Run()
}
