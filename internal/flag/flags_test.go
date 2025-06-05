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
	helpErr, ok := err.(*HelpRequested)
	assert.True(t, ok)
	assert.Contains(t, helpErr.Error(), "Usage: go-snapraid [flags]")
}

func TestParseFlags_Version(t *testing.T) {
	t.Parallel()

	_, err := ParseFlags([]string{"--version"}, "v9.8.7")
	assert.Error(t, err)
	verErr, ok := err.(*HelpRequested)
	assert.True(t, ok)
	assert.Contains(t, verErr.Error(), "go-snapraid version v9.8.7")
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

func TestOptions_Validate(t *testing.T) {
	t.Parallel()

	t.Run("Valid JSON", func(t *testing.T) {
		opts := Options{LogFormat: "json"}
		assert.NoError(t, opts.Validate())
	})

	t.Run("Valid Text", func(t *testing.T) {
		opts := Options{LogFormat: "text"}
		assert.NoError(t, opts.Validate())
	})

	t.Run("Invalid Format", func(t *testing.T) {
		opts := Options{LogFormat: "xml"}
		err := opts.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid log format")
	})
}
