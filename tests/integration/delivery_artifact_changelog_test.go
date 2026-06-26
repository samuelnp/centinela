package integration_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/delivery"
)

const realisticChangelog = `# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- prior feature

### Changed

### Fixed

- prior fix

---

## [0.1.0] - 2026-01-01

### Added

- initial release
`

// TestInsertEntryRoundTripStable drives InsertEntry against a realistic
// multi-section CHANGELOG and asserts a second insert is a no-op.
func TestInsertEntryRoundTripStable(t *testing.T) {
	entry := delivery.ChangelogEntry{Category: "Added", Line: "feat: delivery artifacts"}

	out, changed := delivery.InsertEntry(realisticChangelog, entry)
	if !changed {
		t.Fatal("first insert should change the changelog")
	}
	if strings.Count(out, "- feat: delivery artifacts") != 1 {
		t.Fatalf("expected exactly one new bullet:\n%s", out)
	}
	// The new line must live inside the [Unreleased] block, above the --- rule.
	rule := strings.Index(out, "\n---")
	if idx := strings.Index(out, "- feat: delivery artifacts"); idx < 0 || idx > rule {
		t.Fatalf("new bullet leaked out of [Unreleased]:\n%s", out)
	}
	// Released sections untouched: "initial release" still present exactly once.
	if strings.Count(out, "- initial release") != 1 {
		t.Fatalf("released section was modified:\n%s", out)
	}

	out2, changed2 := delivery.InsertEntry(out, entry)
	if changed2 || out2 != out {
		t.Fatalf("second insert must be an idempotent no-op")
	}
}
