package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// auditOversizedBody returns a >100-line Go body so file_size fails for it.
func auditOversizedBody() string {
	var b strings.Builder
	b.WriteString("package big\n")
	for i := 0; i < 115; i++ {
		b.WriteString("// filler line to exceed the 100-line file-size gate limit\n")
	}
	return b.String()
}

// auditRepo chdirs into a temp repo with file_size + audit_baseline enabled and
// one oversized internal file, reverting chdir on cleanup.
func auditRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	toml := "[gates]\nfile_size = true\n\n[gates.audit_baseline]\nenabled = true\nseverity = \"fail\"\n"
	writeAudit(t, dir, "centinela.toml", toml)
	writeAudit(t, dir, "internal/big.go", auditOversizedBody())
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeAudit(t *testing.T, dir, name, body string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// runCmd invokes fn with a fresh cobra command wired to a buffer, returning the
// captured output and the returned error.
func runCmd(t *testing.T, asJSON bool, fn func(*cobra.Command, []string) error) (string, error) {
	t.Helper()
	old := auditJSON
	auditJSON = asJSON
	t.Cleanup(func() { auditJSON = old })
	c := &cobra.Command{}
	var buf bytes.Buffer
	c.SetOut(&buf)
	err := fn(c, nil)
	return buf.String(), err
}
