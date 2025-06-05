package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gi8lino/go-snapraid/pkg/snapraid"
	"github.com/stretchr/testify/assert"
)

func TestPartialHandler(t *testing.T) {
	t.Parallel()

	fs := fstest.MapFS{
		"web/templates/overview.html": &fstest.MapFile{Data: []byte(`{{define "overview"}}OK{{end}}`)},
		"web/templates/run.html":      &fstest.MapFile{Data: []byte(`{{define "run"}}RUN{{end}}`)},
	}

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	t.Run("Renders template", func(t *testing.T) {
		t.Parallel()

		now := time.Now().UTC().Truncate(time.Second)
		run := snapraid.RunResult{
			Timestamp: now.Format("2006-01-02 15:04"),
			Result: snapraid.DiffResult{
				Added:    []string{"a"},
				Removed:  []string{"b"},
				Updated:  []string{"c"},
				Moved:    []string{"d"},
				Copied:   []string{"e"},
				Restored: []string{"f"},
			},
			Timings: snapraid.RunTimings{
				Touch: 1 * time.Second,
				Diff:  2 * time.Second,
				Sync:  3 * time.Second,
				Scrub: 4 * time.Second,
				Smart: 5 * time.Second,
				Total: 6 * time.Second,
			},
		}
		data := encodeJSON(t, run)

		tmp := t.TempDir()
		jsonPath := filepath.Join(tmp, now.Format(time.RFC3339)+".json")
		assert.NoError(t, os.WriteFile(jsonPath, data, 0o600))

		handler := PartialHandler(fs, tmp, logger)

		req := httptest.NewRequest("GET", "/partials/overview", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "OK")
	})

	t.Run("Run not found", func(t *testing.T) {
		t.Parallel()

		handler := PartialHandler(fs, t.TempDir(), logger)

		req := httptest.NewRequest("GET", "/partials/run?id=nonexistent", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Invalid path", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/doesnotexist", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(fs, t.TempDir(), logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func encodeJSON(t *testing.T, val any) []byte {
	t.Helper()
	var buf bytes.Buffer
	assert.NoError(t, json.NewEncoder(&buf).Encode(val))
	return buf.Bytes()
}
