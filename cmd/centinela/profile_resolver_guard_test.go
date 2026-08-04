package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// profileSurfaces are the files that RESOLVE or REPORT an enforcement profile.
// Each must obtain its config from config.LoadForProfile, which fails closed to
// strict, and none may call config.Load directly.
var profileSurfaces = []string{
	"status_model.go",            // centinela status
	"mcp_rules.go",               // MCP read_rules
	"hook_setup_cascade.go",      // greenfield setup cascade weight
	"roadmap_promote_cascade.go", // promote's grading requirement
	"hook_autostart.go",          // pins profileContract for life
	"hook_context.go",            // step-confirmation directive
	"hook_prewrite.go",           // prewrite step gating
	"hook_plan_advisor.go",       // plan advisor cadence
	"hook_statusline_view.go",    // status-line profile block
}

// guardedTrees are the package directories the sweep parses. cmd/centinela holds
// most surfaces; internal/doctor reports a profile to the operator and sat
// outside the original glob, which is how its fail-open shipped unseen.
//
// NOT COVERED, plainly: packages outside this list, a config obtained via a
// helper that itself calls config.Load, and any idiom beyond the two in
// configSwallow. The load-bearing guarantee is NOT this scan — it is
// config.ResolvedByLoad, which makes every unloaded config resolve strict at the
// decision. Defense in depth, not the defense.
var guardedTrees = []string{".", filepath.Join("..", "..", "internal", "doctor")}

// loadEscapeAllowlist names the ONE file that may swallow a config.Load error:
// hook_migrate.go loads config only to resolve the local provider for a setup
// sync plan and makes no profile decision. An entry here claims the file decides
// nothing about enforcement — check that before adding one.
var loadEscapeAllowlist = map[string]bool{"hook_migrate.go": true}

func readSource(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(body)
}

// TestProfileSurfacesUseTheSharedResolver: fixing fail-open one call site at a
// time is whack-a-mole. Every profile surface routes through LoadForProfile so
// it inherits the fail-closed direction instead of choosing it.
func TestProfileSurfacesUseTheSharedResolver(t *testing.T) {
	for _, name := range profileSurfaces {
		src := readSource(t, name)
		if !strings.Contains(src, "config.LoadForProfile(") {
			t.Errorf("%s resolves a profile but does not use config.LoadForProfile", name)
		}
		if strings.Contains(src, "config.Load()") {
			t.Errorf("%s must not call config.Load directly — use config.LoadForProfile", name)
		}
	}
}

// TestNoSwallowedConfigLoadOutsideTheAllowlist sweeps every guarded package for
// BOTH shipped swallow idioms — the discard and the rebind-to-empty. The rebind
// is the one the original guard missed, and it is the form four hooks used.
func TestNoSwallowedConfigLoadOutsideTheAllowlist(t *testing.T) {
	for _, dir := range guardedTrees {
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatalf("glob %s: %v", dir, err)
		}
		for _, path := range files {
			name := filepath.Base(path)
			if strings.HasSuffix(name, "_test.go") || loadEscapeAllowlist[name] {
				continue
			}
			for _, s := range scanConfigSwallows(t, path) {
				t.Errorf("%s: %s swallows a config.Load failure (discards=%v rebinds=%v); "+
					"use config.LoadForProfile, whose fallback pins strict",
					path, s.Func, s.Discards, s.Rebinds)
			}
		}
	}
}
