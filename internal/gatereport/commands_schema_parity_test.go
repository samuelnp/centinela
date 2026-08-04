package gatereport

import (
	"strings"
	"testing"
)

// commandsReport wraps a raw commands array in a stamped-looking block.
func commandsReport(commands string) string {
	return blockReport(`{"revision":"a","treeDigest":"b","commands":` + commands + `}`)
}

// TestAssessRejectsNullExitCodeRecord guards the actual laundering path, not
// just the shape rule: a report whose ONLY command is a null-exit-code
// `centinela validate` used to be reported as grounded, so a verifier whose
// validate run exited non-zero could record it as a pass.
func TestAssessRejectsNullExitCodeRecord(t *testing.T) {
	report := commandsReport(`[{"argv":["centinela","validate"],"exitCode":null}]`)
	if err := Assess(report); err == nil {
		t.Fatal("a null exitCode must never produce a grounded verdict")
	}
	if _, err := ParseVerification(report); err == nil {
		t.Fatal("the read path must refuse the same record")
	}
	if _, err := Stamped(report, "rev", "sha256:d"); err == nil {
		t.Fatal("the write path must refuse the same record")
	}
}

// TestCommandsSchemaReadWriteParity locks the "one shape, one place" claim:
// Stamped and ParseVerification must refuse the same records with the same
// words, or the two sides can drift the moment either grows a branch.
func TestCommandsSchemaReadWriteParity(t *testing.T) {
	for _, commands := range []string{
		`[{"argv":["centinela","validate"],"exitCode":null}]`,
		`[{"argv":["centinela","validate"]}]`,
		`[{"argv":[],"exitCode":0}]`,
		`[{"argv":"centinela validate","exitCode":0}]`,
		`["centinela validate"]`,
		`[{"argv":["c","validate"],"exitCode":0,"durationMs":null}]`,
	} {
		report := commandsReport(commands)
		_, readErr := ParseVerification(report)
		_, writeErr := Stamped(report, "rev", "sha256:d")
		if readErr == nil || writeErr == nil {
			t.Fatalf("%s: read=%v write=%v — both sides must refuse", commands, readErr, writeErr)
		}
		if readErr.Error() != writeErr.Error() {
			t.Errorf("%s: read %q != write %q", commands, readErr, writeErr)
		}
	}
}

// A record hand-edited AFTER stamping never re-enters the write path, so the
// reader must be exactly as strict as the writer.
func TestStampedBlockStaysBindingAfterHandEdit(t *testing.T) {
	stamped, err := Stamped(commandsReport(`[{"argv":["centinela","validate"],"exitCode":0}]`), "rev", "sha256:d")
	if err != nil {
		t.Fatalf("well-formed record must stamp: %v", err)
	}
	if err := Assess(stamped); err != nil {
		t.Fatalf("stamped grounded report rejected: %v", err)
	}
	edited := strings.Replace(stamped, `"exitCode":0`, `"exitCode":null`, 1)
	if edited == stamped {
		t.Fatal("fixture did not apply the hand edit")
	}
	if err := Assess(edited); err == nil {
		t.Fatal("hand-edited null exitCode must be refused on read")
	}
}
