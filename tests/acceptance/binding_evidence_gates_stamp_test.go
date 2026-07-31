// Acceptance: specs/binding-evidence-gates.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// begStamp seeds a real git repo with a report carrying commands and runs the
// REAL `centinela artifact stamp` against it. Origin is never a network
// remote — the fixture is a local repo and nothing is pushed.
func begStamp(t *testing.T, feature, commands string) (dir, out string, code int) {
	t.Helper()
	bin := avvBuildBin(t)
	dir = avvFixture(t)
	avvSeedWorkflow(t, dir, feature, "adversarial-v1")
	avvWriteReport(t, dir, feature, avvReport("SAFE", "- none", commands))
	out, code = runCent(t, bin, dir, "artifact", "stamp", feature)
	return dir, out, code
}

// Scenario: A malformed commands array is rejected at stamp time
func TestBEG_StampRejectsEntryMissingExitCode(t *testing.T) {
	_, out, code := begStamp(t, "demo", `[{"argv":["centinela","validate"]}]`)
	if code == 0 {
		t.Fatalf("a commands entry with no exitCode must not stamp: %s", out)
	}
	for _, want := range []string{"commands[0]", "exitCode"} {
		if !strings.Contains(out, want) {
			t.Errorf("stamp refusal must name the malformed entry (%q): %s", want, out)
		}
	}
}

// Scenario: A commands entry with an empty argv is rejected at stamp time
func TestBEG_StampRejectsEmptyArgv(t *testing.T) {
	_, out, code := begStamp(t, "demo", `[{"argv":[],"exitCode":0}]`)
	if code == 0 {
		t.Fatalf("an empty argv must not stamp: %s", out)
	}
	if !strings.Contains(out, "empty argv") {
		t.Fatalf("stamp refusal must name the empty argv: %s", out)
	}
}

// A null exitCode is the shape that DEFEATED the schema check: the key was
// present and the decode "succeeded", so the record landed as exit 0. It is
// refused at write time now, exactly like the absent key beside it.
//
// (Proposed spec scenario "A null exit code is rejected at stamp time" — the
// prewrite hook blocks specs/ edits during the tests step; text is in the
// qa-senior evidence report.)
func TestBEG_StampRejectsNullExitCode(t *testing.T) {
	_, out, code := begStamp(t, "demo", `[{"argv":["centinela","validate"],"exitCode":null}]`)
	if code == 0 {
		t.Fatalf("a null exitCode must not stamp: %s", out)
	}
	if !strings.Contains(out, "exitCode") || !strings.Contains(out, "null") {
		t.Fatalf("stamp refusal must name the null exitCode: %s", out)
	}
}

// Scenario: A well-formed commands array is accepted at stamp time
func TestBEG_StampAcceptsWellFormedCommands(t *testing.T) {
	commands := `[{"argv":["centinela","validate"],"exitCode":0,"durationMs":84210}]`
	dir, out, code := begStamp(t, "demo", commands)
	if code != 0 {
		t.Fatalf("a well-formed record must stamp: %s", out)
	}
	stamped := readFile(t, filepath.Join(dir, ".workflow", "demo-gatekeeper.md"))
	if !strings.Contains(stamped, commands) {
		t.Fatalf("the commands array must be preserved verbatim:\n%s", stamped)
	}
	if strings.Contains(stamped, `"revision": ""`) || strings.Contains(stamped, `"treeDigest": ""`) {
		t.Fatalf("stamp must record a real revision and treeDigest:\n%s", stamped)
	}
}
