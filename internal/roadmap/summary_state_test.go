package roadmap

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestSaveLoadSummaryStateRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	path := SummaryStatePath()
	want := SummaryState{SessionID: "s-1", Digest: "abc123"}
	if err := SaveSummaryState(path, want); err != nil {
		t.Fatalf("save: %v", err)
	}
	if got := LoadSummaryState(path); got != want {
		t.Fatalf("round trip: got %+v want %+v", got, want)
	}
	// The temp file must not survive the rename.
	entries, _ := filepath.Glob(".workflow/.roadmap-digest-*")
	if len(entries) != 0 {
		t.Fatalf("temp files left behind: %v", entries)
	}
}

// E19: absent, corrupt and directory-shaped state all degrade to the zero
// value, which ShouldRenderSummary turns into a render.
func TestLoadSummaryStateDegradesToZeroValue(t *testing.T) {
	t.Chdir(t.TempDir())
	if got := LoadSummaryState(".workflow/.roadmap-digest"); got != (SummaryState{}) {
		t.Fatalf("absent state: got %+v", got)
	}
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".workflow/.roadmap-digest", []byte(`{"sessionId":`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := LoadSummaryState(".workflow/.roadmap-digest"); got != (SummaryState{}) {
		t.Fatalf("corrupt state: got %+v", got)
	}
	if err := os.MkdirAll(".workflow/dirstate", 0o755); err != nil {
		t.Fatal(err)
	}
	if got := LoadSummaryState(".workflow/dirstate"); got != (SummaryState{}) {
		t.Fatalf("unreadable state: got %+v", got)
	}
}

// E20: an unwritable directory returns an error the caller ignores — it must
// never panic and never leave the caller without a render decision.
func TestSaveSummaryStateUnwritableDirReturnsError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(".workflow", 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(dir, ".workflow"), 0o755) })
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	if err := SaveSummaryState(SummaryStatePath(), SummaryState{SessionID: "s"}); err == nil {
		t.Fatal("expected an error for an unwritable .workflow/")
	}
}

// E22: concurrent saves never yield a partial-write corrupt state.
func TestSaveSummaryStateConcurrent(t *testing.T) {
	t.Chdir(t.TempDir())
	path := SummaryStatePath()
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = SaveSummaryState(path, SummaryState{SessionID: "s-1", Digest: "d-1"})
		}()
	}
	wg.Wait()
	if got := LoadSummaryState(path); got.SessionID != "s-1" || got.Digest != "d-1" {
		t.Fatalf("concurrent saves corrupted state: %+v", got)
	}
	entries, _ := filepath.Glob(".workflow/.roadmap-digest-*")
	if len(entries) != 0 {
		t.Fatalf("temp files left behind: %v", entries)
	}
}
