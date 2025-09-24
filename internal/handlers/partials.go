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

// OverviewView represents a summarized SnapRAID run for display in the overview table.
type OverviewView struct {
	Timestamp string        // original RFC3339 timestamp used as run ID
	Date      string        // formatted timestamp for display
	Total     int           // total number of file changes
	TouchTime time.Duration // duration of `touch` step
	DiffTime  time.Duration // duration of `diff` step
	SyncTime  time.Duration // duration of `sync` step
	ScrubTime time.Duration // duration of `scrub` step
	SmartTime time.Duration // duration of `smart` step
	TotalTime time.Duration // total runtime duration
}

// RunView represents detailed file-level changes for a specific SnapRAID run.
type RunView struct {
	Timestamp     string   // run ID / timestamp
	Date          string   // formatted run timestamp
	AddedFiles    []string // list of added files
	RemovedFiles  []string // list of removed files
	UpdatedFiles  []string // list of updated files
	MovedFiles    []string // list of moved files
	CopiedFiles   []string // list of copied files
	RestoredFiles []string // list of restored files
}

type runResultCompat struct {
	Timestamp string              `json:"timestamp"`
	Result    snapraid.RunResult  `json:"result"`
	Timings   snapraid.RunTimings `json:"timings"`
	Error     json.RawMessage     `json:"error"`
}

// notFoundError is returned by the handler when a requested partial section is not found.
type notFoundError struct {
	msg string
}

// Error implements the error interface.
func (e *notFoundError) Error() string { return e.msg }

// PartialHandler returns an HTTP handler that renders HTML templates for partial sections.
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
				"web/templates/run.html",
			),
	)

	return func(w http.ResponseWriter, r *http.Request) {
		section := path.Base(r.URL.Path)
		var err error

		switch section {
		case "overview":
			err = renderOverview(w, tmpl, outputDir)

		case "run":
			runID := r.URL.Query().Get("id")
			if runID == "" {
				runID, err = findLatestRunID(outputDir)
				if err != nil {
					break
				}
			}
			err = renderRun(w, tmpl, outputDir, runID)
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

// renderOverview renders the overview partial with a summary of all runs.
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
	for _, fullPath := range matches {
		timestampStr := strings.TrimSuffix(filepath.Base(fullPath), ".json")
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			continue
		}

		f, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("open file %q failed: %w", fullPath, err)
		}
		defer f.Close() // nolint:errcheck

		var result snapraid.RunResult
		if err := json.NewDecoder(f).Decode(&result); err != nil {
			return fmt.Errorf("JSON decode of %q failed: %w", fullPath, err)
		}

		stats := result.Result
		total := len(stats.Added) + len(stats.Removed) + len(stats.Updated) +
			len(stats.Moved) + len(stats.Copied) + len(stats.Restored)

		rows = append(rows, OverviewView{
			Timestamp: timestampStr,
			Date:      timestamp.Format(time.RFC3339),
			Total:     total,
			TouchTime: result.Timings.Touch,
			DiffTime:  result.Timings.Diff,
			SyncTime:  result.Timings.Sync,
			ScrubTime: result.Timings.Scrub,
			SmartTime: result.Timings.Smart,
			TotalTime: result.Timings.Total,
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

// renderRun renders the detailed view for a single SnapRAID run.
func renderRun(
	w io.Writer,
	tmpl *template.Template,
	outputDir string,
	runID string,
) error {
	fullPath := filepath.Join(outputDir, runID+".json")
	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &notFoundError{fmt.Sprintf("run %q not found", runID)}
		}
		return fmt.Errorf("open detail file %q failed: %w", fullPath, err)
	}
	defer f.Close() // nolint:errcheck

	var result runResultCompat
	if err := json.NewDecoder(f).Decode(&result); err != nil {
		return fmt.Errorf("JSON decode %q failed: %w", fullPath, err)
	}

	// build dropdown list
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

	return tmpl.ExecuteTemplate(w, "run", struct {
		Run           RunView
		AllTimestamps []string
	}{
		Run: RunView{
			Timestamp:     runID,
			Date:          result.Timestamp,
			AddedFiles:    result.Result.Added,
			RemovedFiles:  result.Result.Removed,
			UpdatedFiles:  result.Result.Updated,
			MovedFiles:    result.Result.Moved,
			CopiedFiles:   result.Result.Copied,
			RestoredFiles: result.Result.Restored,
		},
		AllTimestamps: allTimestamps,
	})
}

// findLatestRunID returns the most recent run file's ID from the output directory.
func findLatestRunID(outputDir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("no run files found or glob failed: %w", err)
	}
	slices.Sort(matches) // RFC3339 file names sort lexicographically by time
	latest := filepath.Base(matches[len(matches)-1])
	return strings.TrimSuffix(latest, ".json"), nil
}
