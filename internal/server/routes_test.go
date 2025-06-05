package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()

	// in-memory file system with minimal template files
	webFS := fstest.MapFS{
		"web/static/css/go-snapraid.css": &fstest.MapFile{Data: []byte("body { background: white; }")},
		"web/templates/base.html":        &fstest.MapFile{Data: []byte(`{{define "base"}}<html>{{template "navbar"}}<footer>{{.Version}}</footer>{{end}}`)},
		"web/templates/navbar.html":      &fstest.MapFile{Data: []byte(`{{define "navbar"}}<nav>nav</nav>{{end}}`)},
		"web/templates/overview.html":    &fstest.MapFile{Data: []byte(` {{ define "overview" }}<div id="overview">Overview page</div>{{ end }}`)},
		"web/templates/run.html":         &fstest.MapFile{Data: []byte(` {{ define "run" }}<div id="run">Run page</div>{{ end }}`)},
		"web/templates/footer.html":      &fstest.MapFile{Data: []byte(`{{define "footer"}}<!-- footer -->{{end}}`)},
	}

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	router := NewRouter(webFS, "/does-not-matter", "test-version", logger)

	t.Run("GET /static/css/go-snapraid.css", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/static/css/go-snapraid.css", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "background")
	})

	t.Run("GET /healthz", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("GET / (Home)", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "test-version")
	})

	t.Run("GET /partials/overview", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/partials/overview", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusNotFound, rec.Code)
	})

	t.Run("GET /partials/run", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/partials/run", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusNotFound, rec.Code)
	})
}
