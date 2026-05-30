package memory

import (
	"os"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

func enabledCfg() *config.Config {
	enabled := true
	return &config.Config{
		Memory: config.MemoryConfig{
			Enabled:          &enabled,
			RecallMaxEntries: 10,
			RecallMaxBytes:   4096,
		},
	}
}

func disabledCfg() *config.Config {
	disabled := false
	return &config.Config{Memory: config.MemoryConfig{Enabled: &disabled}}
}

// SC-10: empty ledger returns empty slice, no error.
func TestRecallEmptyLedger(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	result := Recall(Query{Feature: "beta"}, enabledCfg())
	if len(result) != 0 {
		t.Fatalf("expected empty recall on empty ledger, got %d entries", len(result))
	}
}

// SC-12: disabled config → nil result.
func TestRecallDisabledConfig(t *testing.T) {
	result := Recall(Query{Feature: "beta"}, disabledCfg())
	if result != nil {
		t.Fatal("expected nil result when memory disabled (SC-12)")
	}
}

// SC-12: nil config → nil result.
func TestRecallNilConfig(t *testing.T) {
	if Recall(Query{Feature: "beta"}, nil) != nil {
		t.Fatal("expected nil result for nil config")
	}
}

// SC-09: dependency-feature match beats shared tags beats recency.
func TestRecallRankingOrder(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

	eDep := newEntry("dep-feat", "tests", TypeLesson, "dep lesson", "s", []string{"coverage"}, t1)
	eOtherA := newEntry("other-a", "tests", TypeLesson, "other-a lesson", "s", []string{"coverage"}, t2)
	eOtherB := newEntry("other-b", "validate", TypeVerdict, "other-b verdict", "s", []string{"unrelated"}, t3)

	writeIfAbsent(eDep)    //nolint:errcheck
	writeIfAbsent(eOtherA) //nolint:errcheck
	writeIfAbsent(eOtherB) //nolint:errcheck

	q := Query{Feature: "beta", Dependencies: []string{"dep-feat"}, Tags: []string{"coverage"}}
	result := Recall(q, enabledCfg())

	if len(result) < 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
	if result[0].Feature != "dep-feat" {
		t.Fatalf("dep-feat must rank first (SC-09), got %q", result[0].Feature)
	}
	if result[1].Feature != "other-a" {
		t.Fatalf("other-a must rank second (shared tag), got %q", result[1].Feature)
	}
	if result[2].Feature != "other-b" {
		t.Fatalf("other-b must rank last, got %q", result[2].Feature)
	}
}
