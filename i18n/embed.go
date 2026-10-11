package i18n

import (
	"embed"
	"io/fs"
)

//go:embed *.json
var localeFS embed.FS

// EmbeddedLocales is an embed.FS containing all bundled locale JSON files.
// Use this in production builds so that translations work regardless of
// the working directory.
var EmbeddedLocales fs.FS = localeFS
