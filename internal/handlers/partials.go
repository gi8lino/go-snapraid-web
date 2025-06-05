package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/gi8lino/go-snapraid-web/internal/utils"
	"github.com/gi8lino/go-snapraid/pkg/snapraid"
)

type OverviewView struct {
	Timestamp string
	Date      string
	Total     int
	TouchTime time.Duration
	DiffTime  time.Duration
	SyncTime  time.Duration
	ScrubTime time.Duration
	SmartTime time.Duration
	TotalTime time.Duration
}

type DetailsView struct {
	Timestamp     string
	Date          string
	AddedFiles    []string
	RemovedFiles  []string
	UpdatedFiles  []string
	MovedFiles    []string
	CopiedFiles   []string
	RestoredFiles []string
}

type notFoundError struct {
	msg string
}

func (e *notFoundError) Error() string { return e.msg }

func PartialHandler(
	webFS fs.FS,
	outputDir string,
	logger *slog.Logger,
) http.HandlerFunc {
	tmpl := template.Must(
		template.New("partials").
			Funcs(utils.FuncMap()).
			ParseFS(
				webFS,
				"web/templates/overview.html",
				"web/templates/details.html",
			),
	)

	return func(w http.ResponseWriter, r *http.Request) {
		section := path.Base(r.URL.Path)
		var err error
		switch section {
		case "overview":
			err = renderOverview(w, tmpl, outputDir)

		case "details":
			id := r.URL.Query().Get("id")
			if id == "" {
				id, err = findLatestRunID(outputDir)
				if err != nil {
					break
				}
			}
			err = renderDetails(w, tmpl, outputDir, id)
			if errors.As(err, new(*notFoundError)) {
				logger.Error("no run files found or glob failed", "error", err)
				http.NotFound(w, r)
				return
			}
		default:
			http.NotFound(w, r)
		}

		if err != nil {
			logger.Error("render "+section+" partial", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

func renderOverview(
	w io.Writer,
	tmpl *template.Template,
	outputDir string,
) error {
	matches, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err != nil {
		return fmt.Errorf("glob failed: %w", err)
	}

	var rows []OverviewView
	for _, fullpath := range matches {
		ts := strings.TrimSuffix(filepath.Base(fullpath), ".json")
		dt, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			continue
		}

		f, err := os.Open(fullpath)
		if err != nil {
			return fmt.Errorf("open file %q failed: %w", fullpath, err)
		}
		defer f.Close() // nolint:errcheck

		var rr snapraid.RunResult
		if err := json.NewDecoder(f).Decode(&rr); err != nil {
			return fmt.Errorf("JSON decode of %q failed: %w", fullpath, err)
		}

		sum := rr.Result
		totalChanges := len(sum.Added) + len(sum.Removed) + len(sum.Updated) +
			len(sum.Moved) + len(sum.Copied) + len(sum.Restored)

		rows = append(rows, OverviewView{
			Timestamp: ts,
			Date:      dt.Format("2006-01-02 15:04"),
			Total:     totalChanges,
			TouchTime: rr.Timings.Touch,
			DiffTime:  rr.Timings.Diff,
			SyncTime:  rr.Timings.Sync,
			ScrubTime: rr.Timings.Scrub,
			SmartTime: rr.Timings.Smart,
			TotalTime: rr.Timings.Total,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Timestamp > rows[j].Timestamp
	})

	return tmpl.ExecuteTemplate(w, "overview", struct {
		Rows []OverviewView
	}{
		Rows: rows,
	})
}

func renderDetails(
	w io.Writer,
	tmpl *template.Template,
	outputDir string,
	id string,
) error {
	fullpath := filepath.Join(outputDir, id+".json")
	f, err := os.Open(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			return &notFoundError{fmt.Sprintf("run %q not found", id)}
		}
		return fmt.Errorf("open detail file %q failed: %w", fullpath, err)
	}
	defer f.Close() // nolint:errcheck

	var rr snapraid.RunResult
	if err := json.NewDecoder(f).Decode(&rr); err != nil {
		return fmt.Errorf("JSON decode %q failed: %w", fullpath, err)
	}

	// Build dropdown list
	matches, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err != nil {
		return fmt.Errorf("glob for dropdown failed: %w", err)
	}
	var allTimestamps []string
	for _, file := range matches {
		ts := strings.TrimSuffix(filepath.Base(file), ".json")
		allTimestamps = append(allTimestamps, ts)
	}

	slices.Sort(allTimestamps)

	return tmpl.ExecuteTemplate(w, "details", struct {
		Detail        DetailsView
		AllTimestamps []string
	}{
		Detail: DetailsView{
			Timestamp:     id,
			Date:          rr.Timestamp,
			AddedFiles:    rr.Result.Added,
			RemovedFiles:  rr.Result.Removed,
			UpdatedFiles:  rr.Result.Updated,
			MovedFiles:    rr.Result.Moved,
			CopiedFiles:   rr.Result.Copied,
			RestoredFiles: rr.Result.Restored,
		},
		AllTimestamps: allTimestamps,
	})
}

func findLatestRunID(outputDir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("no run files found or glob failed: %w", err)
	}

	slices.Sort(matches)
	latest := filepath.Base(matches[0])
	return strings.TrimSuffix(latest, ".json"), nil
}
