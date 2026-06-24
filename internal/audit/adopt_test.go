package audit

import (
	"os"
	"testing"
)

// TestAdoptFreshRecords: on a repo with a violation and no baseline yet, Adopt
// records and writes a non-empty baseline and reports Skipped=false.
func TestAdoptFreshRecords(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	o, err := Adopt(cfg, false)
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if o.Skipped {
		t.Fatal("fresh adopt should not be skipped")
	}
	if o.Baseline.Total() == 0 {
		t.Fatal("expected at least one baselined violation")
	}
	if _, err := os.Stat(o.Path); err != nil {
		t.Fatalf("baseline file not written at %s: %v", o.Path, err)
	}
}

// TestAdoptSkipsWhenExists: a second Adopt without force is skipped and leaves
// the existing baseline file byte-unchanged.
func TestAdoptSkipsWhenExists(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	first, err := Adopt(cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(first.Path)
	o, err := Adopt(cfg, false)
	if err != nil {
		t.Fatalf("skip path should not error: %v", err)
	}
	if !o.Skipped {
		t.Fatal("second adopt should be skipped")
	}
	after, _ := os.ReadFile(first.Path)
	if string(before) != string(after) {
		t.Fatal("skip must leave the baseline byte-unchanged")
	}
}

// TestAdoptForceOverwrites: with --force a new violation is captured, widening
// the baseline rather than skipping.
func TestAdoptForceOverwrites(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/a.go": oversizedGo(0)})
	if _, err := Adopt(cfg, false); err != nil {
		t.Fatal(err)
	}
	write(t, dirOfAudit(t), "internal/b.go", oversizedGo(3))
	o, err := Adopt(cfg, true)
	if err != nil {
		t.Fatalf("force adopt: %v", err)
	}
	if o.Skipped {
		t.Fatal("force must not skip")
	}
	if o.Baseline.Total() != 2 {
		t.Fatalf("force should capture both violations, got %d", o.Baseline.Total())
	}
}

func dirOfAudit(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}
