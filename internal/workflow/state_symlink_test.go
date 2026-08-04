package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

// A symlinked STATE FILE is replaced, not followed — the accepted cost of
// rename(2) (see WriteFileAtomic). Pinned deliberately so the semantic change
// is a decision on record rather than a surprise: writing through the link
// would re-open the torn-write window, and the link may point across a
// filesystem boundary where rename cannot be atomic at all.
func TestSymlinkedStateFileIsReplacedNotFollowed(t *testing.T) {
	dir := stateRepo(t)
	elsewhere := filepath.Join(dir, "elsewhere.json")
	body := `{"feature":"alpha","currentStep":"code","steps":{}}`
	if err := os.WriteFile(elsewhere, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, FilePath("alpha")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	wf, err := Load("alpha")
	if err != nil {
		t.Fatal(err)
	}
	wf.CurrentStep = "tests"
	if err := Save(wf); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(FilePath("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("the link survived — this test records that rename replaces it")
	}
	after, _ := os.ReadFile(elsewhere)
	if string(after) != body {
		t.Fatalf("the link destination must be left alone, got %q", after)
	}
}

// The supported way to share state between checkouts: symlink the DIRECTORY.
// The temp is then created inside the resolved directory, so the replace stays
// atomic and the shared file really is updated.
func TestSymlinkedStateDirectoryIsWrittenThrough(t *testing.T) {
	dir := stateRepo(t)
	shared := filepath.Join(dir, "shared")
	if err := os.MkdirAll(shared, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(WorkflowDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(shared, WorkflowDir); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := Save(New("alpha")); err != nil {
		t.Fatalf("saving through a symlinked .workflow/ must work: %v", err)
	}
	if _, err := os.Stat(filepath.Join(shared, "alpha.json")); err != nil {
		t.Fatalf("the shared directory did not receive the state file: %v", err)
	}
}
