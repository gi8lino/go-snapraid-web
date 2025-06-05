package logging

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupLogger_JSONFormat(t *testing.T) {
	t.Parallel()
	t.Run("JSON format", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatJSON, &buf)

		logger.Info("json test", "key", "value")
		output := buf.String()

		assert.Contains(t, output, `"msg":"json test"`)
		assert.Contains(t, output, `"key":"value"`)
	})

	t.Run("Text format", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatText, &buf)

		logger.Info("text test", "key", "value")
		output := buf.String()

		assert.Contains(t, output, "text test")
		assert.Contains(t, output, "key=value")
	})

	t.Run("Invalid format defaults to JSON", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger("invalid", &buf)

		logger.Info("default json", "key", "fallback")
		output := buf.String()

		assert.Contains(t, output, `"msg":"default json"`)
		assert.Contains(t, output, `"key":"fallback"`)
	})
}
