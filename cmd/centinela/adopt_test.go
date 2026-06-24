package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runAdoptCmd invokes runAdopt with the given flags against a fresh cobra
// command wired to a buffer, restoring the flag globals on cleanup.
func runAdoptCmd(t *testing.T, force, asJSON bool) (string, error) {
	t.Helper()
	of, oj := adoptForce, adoptJSON
	adoptForce, adoptJSON = force, asJSON
	t.Cleanup(func() { adoptForce, adoptJSON = of, oj })
	c := &cobra.Command{}
	var buf bytes.Buffer
	c.SetOut(&buf)
	err := runAdopt(c, nil)
	return buf.String(), err
}

func adoptBaselineBytes(t *testing.T) []byte {
	t.Helper()
	b, _ := os.ReadFile(filepath.Join(".", ".workflow", "audit-baseline.json"))
	return b
}

// TestRunAdoptHappy records a baseline and prints the adoption report (exit 0).
func TestRunAdoptHappy(t *testing.T) {
	auditRepo(t)
	out, err := runAdoptCmd(t, false, false)
	if err != nil {
		t.Fatalf("adopt errored: %v", err)
	}
	if !strings.Contains(out, "Adopted baseline") || !strings.Contains(out, "ratchet to zero") {
		t.Fatalf("missing adoption report: %q", out)
	}
}

// TestRunAdoptSkipExits non-zero and leaves the file byte-unchanged.
func TestRunAdoptSkip(t *testing.T) {
	auditRepo(t)
	if _, err := runAdoptCmd(t, false, false); err != nil {
		t.Fatal(err)
	}
	before := adoptBaselineBytes(t)
	_, err := runAdoptCmd(t, false, false)
	if err == nil || !strings.Contains(err.Error(), "use --force") {
		t.Fatalf("skip should error with use --force, got %v", err)
	}
	if string(before) != string(adoptBaselineBytes(t)) {
		t.Fatal("skip changed the baseline file")
	}
}

// TestRunAdoptForce overwrites an existing baseline (exit 0).
func TestRunAdoptForce(t *testing.T) {
	auditRepo(t)
	if _, err := runAdoptCmd(t, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := runAdoptCmd(t, true, false); err != nil {
		t.Fatalf("force adopt errored: %v", err)
	}
}

// TestRunAdoptConfigError surfaces a config-load failure.
func TestRunAdoptConfigError(t *testing.T) {
	dir := auditRepo(t)
	writeAudit(t, dir, "centinela.toml", "this = = bad toml")
	if _, err := runAdoptCmd(t, false, false); err == nil {
		t.Fatal("runAdopt should propagate config error")
	}
}
