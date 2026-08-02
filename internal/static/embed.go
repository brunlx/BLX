// Package static embeds the frontend single-page application into the binary.
package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:web
var webFS embed.FS

// Handler serves the embedded frontend assets.
func Handler() (http.Handler, error) {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(sub)), nil
}
