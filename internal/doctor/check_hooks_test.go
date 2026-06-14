package doctor

import (
	"os"
	"strings"
	"testing"
)

func TestHooksCheckNoClaudeDirWarns(t *testing.T) {
	repoFixture(t)
	d := hooksCheck{}.Run(Context{})
	if d.Status != Warn {
		t.Fatalf("no .claude/ must Warn, got %v", d.Status)
	}
	if !strings.Contains(d.Message, "centinela setup") {
		t.Fatalf("message must mention setup: %q", d.Message)
	}
	if d.Repair != nil {
		t.Fatal("warn-degrade must not carry a repair")
	}
}

func TestHooksCheckMissingEntriesError(t *testing.T) {
	repoFixture(t)
	writeFile(t, ".claude/settings.json", "{}")
	d := hooksCheck{}.Run(Context{})
	if d.Status != Error {
		t.Fatalf("missing entries must Error, got %v", d.Status)
	}
	if d.Repair == nil || !d.Repair.Safe || d.Repair.Apply == nil {
		t.Fatal("error path must carry a safe, applicable repair")
	}
	if len(d.Details) == 0 {
		t.Fatal("missing entries should be itemized in details")
	}
}

func TestHooksCheckWiredOK(t *testing.T) {
	repoFixture(t)
	seedSyncedHooks(t)
	d := hooksCheck{}.Run(Context{})
	if d.Status != OK {
		t.Fatalf("fully wired must be OK, got %v (%s)", d.Status, d.Message)
	}
}

func TestHooksRepairFixesAndIsIdempotent(t *testing.T) {
	repoFixture(t)
	writeFile(t, ".claude/settings.json", "{}")
	d := hooksCheck{}.Run(Context{})
	if err := d.Repair.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	got := hooksCheck{}.Run(Context{})
	if got.Status != OK {
		t.Fatalf("post-repair must be OK, got %v (%s)", got.Status, got.Message)
	}
	before, _ := os.ReadFile(".claude/settings.json")
	if err := d.Repair.Apply(); err != nil {
		t.Fatalf("idempotent apply: %v", err)
	}
	again, _ := os.ReadFile(".claude/settings.json")
	if string(before) != string(again) {
		t.Fatal("second repair must be byte-identical (idempotent)")
	}
}
