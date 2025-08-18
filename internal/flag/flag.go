package flag

import (
	"net"

	"github.com/gi8lino/go-snapraid-web/internal/logging"

	"github.com/containeroo/tinyflags"
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
	opts := Options{}
	tf := tinyflags.NewFlagSet("go-snapraid", tinyflags.ContinueOnError)
	tf.Version(version)

	// Flags
	listenAddr := tf.TCPAddr("listen-address", &net.TCPAddr{IP: nil, Port: 8080}, "Listen address").
		Short("a").
		Placeholder("ADDR").
		Value()

	tf.StringVar(&opts.OutputDir, "output-dir", "/output", "Output directory for generated files").
		Short("o").
		Value()
	logFormat := tf.String("log-format", "json", "Log format").
		Choices(string(logging.LogFormatText), string(logging.LogFormatJSON)).
		Short("l").
		Value()

	if err := tf.Parse(args); err != nil {
		return Options{}, err
	}

	opts.LogFormat = logging.LogFormat(*logFormat)
	opts.ListenAddr = (*listenAddr).String()

	return opts, nil
}
