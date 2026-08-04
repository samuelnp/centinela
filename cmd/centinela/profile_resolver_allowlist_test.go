package main

import (
	"strings"
	"testing"
)

// TestLoadEscapeAllowlistStaysHonest: an allowlisted file that stops swallowing
// should leave the list, and no allowlisted file may touch profile resolution.
func TestLoadEscapeAllowlistStaysHonest(t *testing.T) {
	for name := range loadEscapeAllowlist {
		if len(scanConfigSwallows(t, name)) == 0 {
			t.Errorf("%s no longer swallows a config.Load failure — drop it from the allowlist", name)
		}
		src := readSource(t, name)
		if strings.Contains(src, "EnforcementProfile") || strings.Contains(src, "ProfileDefaults") {
			t.Errorf("%s touches profile resolution and must not be allowlisted", name)
		}
	}
}
