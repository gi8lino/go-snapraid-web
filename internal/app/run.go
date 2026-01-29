package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os/signal"
	"syscall"

	"github.com/gi8lino/go-snapraid-web/internal/flag"
	"github.com/gi8lino/go-snapraid-web/internal/logging"
	"github.com/gi8lino/go-snapraid-web/internal/server"

	"github.com/containeroo/tinyflags"
)

// Run is the single entry point for the application.
func Run(ctx context.Context, webFS fs.FS, version, commit string, args []string, w io.Writer) error {
	// Parse and validate command-line flags.
	flags, err := flag.ParseFlags(args, version)

	// Setup logger immediately so startup errors are correctly logged.
	logger := logging.SetupLogger(flags.LogFormat, w)
	logger.Info("Starting go-snapraid-web",
		"version", version,
		"commit", commit,
	)

	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			_, _ = fmt.Fprintf(w, "%s\n", err)
			return nil
		}
		logger.Error("Failed to parse CLI flags", "error", err)
		return err
	}

	// Create a context to listen for shutdown signals
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create server and run forever
	router := server.NewRouter(
		webFS,
		flags.OutputDir,
		version,
		logger,
	)
	if err := server.Run(ctx, flags.ListenAddr, router, logger); err != nil {
		logger.Error("Failed to run go-snapraid-web", "error", err)
		return err
	}

	return nil
}
