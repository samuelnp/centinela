package orchestration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadHandoffToReadsTheField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "e.json")
	if err := os.WriteFile(path, []byte(`{"role":"qa-senior","handoffTo":"gatekeeper"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadHandoffTo(path)
	if err != nil || got != "gatekeeper" {
		t.Fatalf("ReadHandoffTo = %q, %v", got, err)
	}
}

// A missing or unparseable file is an ERROR the chain check is expected to
// ignore, not a silent empty string: ValidateRoles already names those, and
// collapsing them to "" here would make an absent file indistinguishable from
// an unset field.
func TestReadHandoffToReportsUnreadableFiles(t *testing.T) {
	dir := t.TempDir()
	if _, err := ReadHandoffTo(filepath.Join(dir, "absent.json")); err == nil {
		t.Fatal("a missing evidence file must report an error")
	}
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadHandoffTo(bad); err == nil {
		t.Fatal("unparseable evidence must report an error")
	}
}

// The merge steward is out-of-band — it never appears in a workflow's ordered
// steps, so it has no derivable successor and its rule is a closed literal
// pair. Every other role's handoff is derived in internal/workflow.
func TestStewardHandoffLiterals(t *testing.T) {
	for value, accept := range map[string]bool{
		"complete": true, "user": true,
		"orchestrator": false, "COMPLETE": false, " user": false, "banana": false,
	} {
		err := validateStewardHandoff("p.json", RoleMergeSteward, value)
		if accept && err != nil {
			t.Errorf("steward handoffTo %q must be accepted: %v", value, err)
		}
		if !accept && err == nil {
			t.Errorf("steward handoffTo %q must be refused", value)
		}
	}
	if err := validateStewardHandoff("p.json", RoleQASeniorEngineer, "gatekeeper"); err != nil {
		t.Fatalf("the literal pair applies to the steward only: %v", err)
	}
}
