package workflow

import (
	"os"
	"strings"
	"testing"
)

// TestDeleteBetweenLoadAndSaveIsAConflict pins R3: a workflow deleted by
// another process — an abandoned feature, a worktree teardown, a `git clean`,
// a `git checkout` landing during a minutes-long `complete` — must not be
// silently resurrected by the save that follows. Before checkNotDeleted this
// returned nil and recreated the file with stale content.
func TestDeleteBetweenLoadAndSaveIsAConflict(t *testing.T) {
	stateRepo(t)
	if err := Save(New("alpha")); err != nil {
		t.Fatal(err)
	}
	wf, err := Load("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(FilePath("alpha")); err != nil {
		t.Fatal(err)
	}

	wf.CurrentStep = "docs"
	err = Save(wf)
	if err == nil {
		t.Fatal("saving a loaded workflow whose file was deleted must be refused")
	}
	for _, want := range []string{FilePath("alpha"), "deleted since this command read it", "centinela start alpha"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("conflict error must contain %q, got: %v", want, err)
		}
	}
	if _, err := os.Stat(FilePath("alpha")); !os.IsNotExist(err) {
		t.Fatalf("the deleted state file was resurrected anyway (%v)", err)
	}
}

// The exemption that keeps `start` and hook_autostart working is unchanged: a
// workflow nobody read carries no digest, so a missing target is genuinely its
// first write.
func TestNeverLoadedWorkflowStillCreatesAMissingFile(t *testing.T) {
	stateRepo(t)
	if err := Save(New("zeta")); err != nil {
		t.Fatalf("a never-loaded workflow must create its state file: %v", err)
	}
	if _, err := os.Stat(FilePath("zeta")); err != nil {
		t.Fatalf("state file was not created: %v", err)
	}
}

// A workflow that was loaded, saved, and then deleted is still a conflict: the
// digest Save refreshed makes it a LOADED workflow, not a fresh one.
func TestDeleteAfterSaveIsAlsoAConflict(t *testing.T) {
	stateRepo(t)
	wf := New("alpha")
	if err := Save(wf); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(FilePath("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := Save(wf); err == nil {
		t.Fatal("a workflow that published bytes and lost them must report a conflict")
	}
}
