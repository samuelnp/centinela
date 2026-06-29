package main

import (
	"os"
	"strings"
	"testing"
)

// TestRunHookSetupBrownfieldDirective: an initialized repo (centinela.toml) with
// source (go.mod) but no PROJECT.md routes to the brownfield setup directive.
func TestRunHookSetupBrownfieldDirective(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                  //nolint:errcheck
	os.Chdir(d)                                        //nolint:errcheck
	os.WriteFile("centinela.toml", []byte("x"), 0644)  //nolint:errcheck
	os.WriteFile("go.mod", []byte("module x\n"), 0644) //nolint:errcheck

	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "CENTINELA DIRECTIVE: brownfield setup") {
			t.Fatalf("expected brownfield directive, got %q", out)
		}
		if !strings.Contains(out, "BROWNFIELD PROJECT DETECTED") {
			t.Fatalf("expected brownfield panel content, got %q", out)
		}
	})
}

// TestRunHookSetupGreenfieldUnchanged: an initialized but empty repo (no source)
// keeps the existing question-based greenfield directive.
func TestRunHookSetupGreenfieldUnchanged(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                 //nolint:errcheck
	os.Chdir(d)                                       //nolint:errcheck
	os.WriteFile("centinela.toml", []byte("x"), 0644) //nolint:errcheck

	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.HasPrefix(out, "CENTINELA DIRECTIVE: setup required") {
			t.Fatalf("expected greenfield setup directive, got %q", out)
		}
		if strings.Contains(out, "BROWNFIELD") {
			t.Fatalf("greenfield output must not mention BROWNFIELD, got %q", out)
		}
	})
}
