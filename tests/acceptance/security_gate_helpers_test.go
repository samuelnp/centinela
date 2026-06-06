package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// Shared helpers for the security-gate acceptance tier. They drop fake scanner
// binaries on a controlled PATH and drive the REAL gate through
// gates.RunWithFilter, mirroring the makeFakeBin pattern from the unit tier.

// secBin writes an executable named name with body into dir (creating it).
func secBin(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
}

// secPath points PATH at a fresh temp dir and returns it so a test can drop
// fake scanners that the gate resolves via exec.LookPath.
func secPath(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	t.Setenv("PATH", d)
	return d
}

// secCfg returns a config with the security gate enabled and both families
// configured to the v1 tool names.
func secCfg() *config.Config {
	cfg := &config.Config{}
	cfg.Gates.Security = config.SecurityGateConfig{
		Enabled: true,
		Secrets: config.SecretsConfig{Tool: "gitleaks"},
		Vuln:    config.VulnConfig{Tools: []string{"govulncheck", "osv-scanner"}},
	}
	return cfg
}

// secResult finds the gate result whose name starts with prefix.
func secResult(t *testing.T, results []gates.Result, prefix string) (gates.Result, bool) {
	t.Helper()
	for _, r := range results {
		if strings.HasPrefix(r.Name, prefix) {
			return r, true
		}
	}
	return gates.Result{}, false
}
