package utils

import (
	"strings"
	"text/template"
	"time"
)

// FuncMap returns a set of custom template functions for use in templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"duration": func(s string) time.Duration {
			d, _ := time.ParseDuration(s)
			return d
		},
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
	}
}
