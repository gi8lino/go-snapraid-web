package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFuncMap(t *testing.T) {
	t.Parallel()

	t.Run("duration parses valid string", func(t *testing.T) {
		t.Parallel()

		fn := FuncMap()["duration"].(func(string) time.Duration)
		d := fn("1h30m")
		assert.Equal(t, 90*time.Minute, d)
	})

	t.Run("duration on invalid string returns 0", func(t *testing.T) {
		t.Parallel()

		fn := FuncMap()["duration"].(func(string) time.Duration)
		d := fn("notaduration")
		assert.Equal(t, time.Duration(0), d)
	})

	t.Run("title capitalizes first letter", func(t *testing.T) {
		t.Parallel()

		fn := FuncMap()["title"].(func(string) string)
		assert.Equal(t, "Hello", fn("hello"))
		assert.Equal(t, "Go", fn("go"))
		assert.Equal(t, "123abc", fn("123abc"))
	})

	t.Run("title with empty string returns empty", func(t *testing.T) {
		t.Parallel()

		fn := FuncMap()["title"].(func(string) string)
		assert.Equal(t, "", fn(""))
	})
}
