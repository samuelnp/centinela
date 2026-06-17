package main

import (
	"os"
	"strings"
	"testing"
)

// Scenario 6: the prompt context hook degrades with a warning on a corrupted
// centinela.toml — it exits zero (nil error) so the host session continues,
// and injects a "config warning:" line into the context output.
func TestRunHookContextCorruptConfigWarnsAndExitsZero(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("centinela.toml", []byte("this is = = not toml"), 0644); err != nil {
		t.Fatal(err)
	}

	var ctx string
	out := captureStdout(t, func() {
		withStdin(t, "{}", func() {
			if err := runHookContext(nil, nil); err != nil {
				t.Fatalf("hook must not break the session, got error: %v", err)
			}
		})
	})
	ctx = out
	if !strings.Contains(ctx, "config warning:") {
		t.Fatalf("injected context must contain a config warning, got: %q", ctx)
	}
}
