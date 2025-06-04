package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/gi8lino/go-snapraid-webui/internal/handlers"
)

// NewRouter creates a new HTTP router
func NewRouter(
	webFS fs.FS,
	outputDir string,
	version string,
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// Handler for embedded static files
	staticContent, _ := fs.Sub(webFS, "web/static")
	fileServer := http.FileServer(http.FS(staticContent))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.Handle("/", handlers.HomeHandler(webFS, version)) // no Method allowed, otherwise it crashes
	mux.Handle("GET /partials/", http.StripPrefix("/partials", handlers.PartialHandler(webFS, outputDir, logger)))

	mux.Handle("GET /healthz", handlers.Healthz())

	return mux
}
