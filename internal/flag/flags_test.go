package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpRequested_Error(t *testing.T) {
	t.Parallel()
	err := &HelpRequested{Message: "help wanted"}
	assert.Equal(t, "help wanted", err.Error())
}

func TestParseFlags_Defaults(t *testing.T) {
	t.Parallel()

	opts, err := ParseFlags([]string{}, "v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, ":8080", opts.ListenAddr)
	assert.Equal(t, "/output", opts.OutputDir)
	assert.Equal(t, "json", string(opts.LogFormat))
}

func TestParseFlags_Help(t *testing.T) {
	t.Parallel()

	_, err := ParseFlags([]string{"--help"}, "v1.2.3")
	assert.Error(t, err)
	expected := `Usage: go-snapraid [flags]
Flags:
    -a, --listen-address ADDR     Listen address (Default: :8080)
    -o, --output-dir OUTPUT-DIR   Output directory for generated files (Default: /output)
    -l, --log-format <text|json>  Log format (Allowed: text, json) (Default: json)
    -h, --help                    Show help
        --version                 Show version
`
	assert.EqualError(t, err, expected)
}

func TestParseFlags_Version(t *testing.T) {
	t.Parallel()

	_, err := ParseFlags([]string{"--version"}, "v9.8.7")
	assert.Error(t, err)
	assert.EqualError(t, err, "v9.8.7")
}

func TestParseFlags_CustomValues(t *testing.T) {
	t.Parallel()

	args := []string{
		"--listen-address", "0.0.0.0:9999",
		"--output-dir", "/tmp/snap",
		"--log-format", "text",
	}
	opts, err := ParseFlags(args, "v0.0.1")
	assert.NoError(t, err)
	assert.Equal(t, "0.0.0.0:9999", opts.ListenAddr)
	assert.Equal(t, "/tmp/snap", opts.OutputDir)
	assert.Equal(t, "text", string(opts.LogFormat))
}
