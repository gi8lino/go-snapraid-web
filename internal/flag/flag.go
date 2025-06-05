package flag

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gi8lino/go-snapraid-web/internal/logging"

	flag "github.com/spf13/pflag"
)

// HelpRequested indicates that help was requested.
type HelpRequested struct {
	Message string // Message is the help message.
}

// Error returns the help message.
func (e *HelpRequested) Error() string {
	return e.Message
}

// Options holds the application configuration.
type Options struct {
	LogFormat  logging.LogFormat // Specify the log output format
	ListenAddr string            // Address to listen on
	OutputDir  string            // Output directory for generated files
}

// ParseFlags parses command-line flags.
func ParseFlags(args []string, version string) (Options, error) {
	fs := flag.NewFlagSet("go-snapraid", flag.ContinueOnError)
	fs.SortFlags = false

	// Server settings
	listenAddress := fs.StringP("listen-address", "a", ":8080", "Address to listen on")
	outputDir := fs.StringP("output-dir", "o", "/output", "Output directory for generated files")
	logFormat := fs.StringP("log-format", "l", "json", "Log format (json | text)")

	// Meta
	var showHelp, showVersion bool
	fs.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")

	// Custom usage message.
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags]\n\nFlags:\n", strings.ToLower(fs.Name())) // nolint:errcheck
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}

	if showVersion {
		return Options{}, &HelpRequested{Message: fmt.Sprintf("%s version %s\n", fs.Name(), version)}
	}
	if showHelp {
		// Capture custom usage output into buffer
		var buf bytes.Buffer
		fs.SetOutput(&buf)
		fs.Usage()
		return Options{}, &HelpRequested{Message: buf.String()}
	}

	return Options{
		ListenAddr: *listenAddress,
		LogFormat:  logging.LogFormat(*logFormat),
		OutputDir:  *outputDir,
	}, nil
}

// Validate checks whether the Config is semantically valid.
func (c *Options) Validate() error {
	if c.LogFormat != logging.LogFormatText && c.LogFormat != logging.LogFormatJSON {
		return fmt.Errorf("invalid log format: '%s'", c.LogFormat)
	}
	return nil
}
