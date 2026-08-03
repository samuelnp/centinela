package workflow

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// writeRawState drops hand-authored bytes at .workflow/<feature>.json, the way
// every legacy state file and every existing test fixture looks on disk.
func writeRawState(t *testing.T, feature, body string) {
	t.Helper()
	if err := os.WriteFile(FilePath(feature), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadTreatsAbsentVersionAsOne(t *testing.T) {
	stateRepo(t)
	writeRawState(t, "legacy", `{"feature":"legacy","currentStep":"code","steps":{}}`)
	wf, err := Load("legacy")
	if err != nil {
		t.Fatalf("a versionless legacy file must load: %v", err)
	}
	if wf.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", wf.SchemaVersion)
	}
}

func TestSaveStampsCurrentVersion(t *testing.T) {
	stateRepo(t)
	if err := Save(New("beta")); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(FilePath("beta"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"schemaVersion": `+strconv.Itoa(SchemaVersion)) {
		t.Fatalf("state file is not stamped: %s", raw)
	}
	if onDiskVersion(raw) != SchemaVersion {
		t.Fatalf("onDiskVersion = %d, want %d", onDiskVersion(raw), SchemaVersion)
	}
}

// A future-version file must LOAD (a failing Load empties ActiveWorkflows and
// makes the prewrite hook block every governed write) and be refused on SAVE.
func TestFutureVersionLoadsButRefusesToSave(t *testing.T) {
	stateRepo(t)
	body := `{"schemaVersion":99,"feature":"delta","currentStep":"code","steps":{}}`
	writeRawState(t, "delta", body)

	wf, err := Load("delta")
	if err != nil {
		t.Fatalf("a future-version file must still load: %v", err)
	}
	if wf.SchemaVersion != 99 {
		t.Fatalf("SchemaVersion = %d, want the on-disk 99", wf.SchemaVersion)
	}

	err = Save(wf)
	if err == nil {
		t.Fatal("saving over a newer schema must be refused")
	}
	for _, want := range []string{FilePath("delta"), "99", "schema version 1", "centinela update"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("refusal must name %q, got: %v", want, err)
		}
	}
	after, err := os.ReadFile(FilePath("delta"))
	if err != nil || string(after) != body {
		t.Fatalf("a refused file must be byte-for-byte unchanged, got %q (%v)", after, err)
	}
}

// The version is read from the ON-DISK bytes, not the in-memory struct: a newer
// binary may have upgraded the file between our Load and our Save.
func TestSaveReadsVersionFromDiskNotMemory(t *testing.T) {
	stateRepo(t)
	writeRawState(t, "delta", `{"feature":"delta","currentStep":"code","steps":{}}`)
	wf, err := Load("delta")
	if err != nil {
		t.Fatal(err)
	}
	writeRawState(t, "delta", `{"schemaVersion":99,"feature":"delta","currentStep":"code","steps":{}}`)
	if err := Save(wf); err == nil || !strings.Contains(err.Error(), "newer Centinela") {
		t.Fatalf("an upgrade behind our back must be caught, got %v", err)
	}
}

func TestOnDiskVersionToleratesGarbage(t *testing.T) {
	if got := onDiskVersion([]byte("{not json")); got != legacyVersion {
		t.Fatalf("corrupt bytes must not read as a future version, got %d", got)
	}
}
