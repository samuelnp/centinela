package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGatherEvidenceToleratesMissingSources(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	e, err := gatherEvidence("ghost")
	if err != nil {
		t.Fatalf("all-missing sources should not error: %v", err)
	}
	if e.Brief != "" || e.Plan != "" || e.GatekeeperReport != "" || e.SpecPath != "" {
		t.Fatalf("expected empty evidence, got %+v", e)
	}
	if e.Feature != "ghost" {
		t.Fatalf("feature slug should be set: %q", e.Feature)
	}
}
