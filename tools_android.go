//go:build android

package main

// Blank import pins golang.org/x/mobile (the version fyne's android driver relies
// on) without importing x/mobile/app, which would duplicate the native activity
// symbols provided by fyne's own forked app package (internal/driver/mobile/app).
import _ "golang.org/x/mobile/event/size"
