package flag

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gi8lino/go-snapraid-web/internal/logging"

	flag "github.com/spf13/pflag"
)

// HelpRequested indicates that help was requested via a CLI flag.
type HelpRequested struct {
	Message string // help message to return as error
}

// Error returns the help message as an error string.
func (e *HelpRequested) Error() string {
	return e.Message
}

// Options holds the parsed configuration flags.
type Options struct {
	LogFormat  logging.LogFormat // log format: json or text
	ListenAddr string            // address to listen on (e.g., ":8080")
	OutputDir  string            // directory to read SnapRAID output JSON files
}

// ParseFlags parses command-line arguments into Options.
func ParseFlags(args []string, version string) (Options, error) {
	fs := flag.NewFlagSet("go-snapraid", flag.ContinueOnError)
	fs.SortFlags = false

	// Flags
	listenAddr := fs.StringP("listen-address", "a", ":8080", "Address to listen on")
	outputPath := fs.StringP("output-dir", "o", "/output", "Output directory for generated files")
	logFormat := fs.StringP("log-format", "l", "json", "Log format (json | text)")

	// Meta flags
	var showHelp, showVersion bool
	fs.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")

	// Custom help output
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
		var buf bytes.Buffer
		fs.SetOutput(&buf)
		fs.Usage()
		return Options{}, &HelpRequested{Message: buf.String()}
	}

	return Options{
		ListenAddr: *listenAddr,
		LogFormat:  logging.LogFormat(*logFormat),
		OutputDir:  *outputPath,
	}, nil
}

// Validate returns an error if the log format is invalid.
func (c *Options) Validate() error {
	if c.LogFormat != logging.LogFormatText && c.LogFormat != logging.LogFormatJSON {
		return fmt.Errorf("invalid log format: '%s'", c.LogFormat)
	}
	return nil
}
