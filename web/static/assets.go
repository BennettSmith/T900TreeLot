// Package static embeds generated browser assets so Node.js is never required
// by the deployed application.
package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed app.css passkeys.js
var files embed.FS

func Handler() http.Handler {
	assets, err := fs.Sub(files, ".")
	if err != nil {
		panic("create embedded static filesystem: " + err.Error())
	}
	return http.StripPrefix("/static/", http.FileServer(http.FS(assets)))
}
