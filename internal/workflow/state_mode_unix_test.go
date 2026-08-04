//go:build !windows

package workflow

import (
	"os"
	"syscall"
	"testing"
)

// The mode is a property of Centinela, not of the operator's shell. The
// explicit Chmod bypasses umask on purpose: os.WriteFile's perm argument is
// masked, so under `umask 077` the pre-atomic writer produced a 0600 state file
// that a hook running under another uid could not read. See stateFileMode.
func TestSaveNormalisesModeRegardlessOfUmask(t *testing.T) {
	stateRepo(t)
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)

	if err := Save(New("alpha")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(FilePath("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("mode = %v under umask 077, want 0644 — every hook reads this file",
			info.Mode().Perm())
	}
}

// A read-only state file is writable again, and normalised: rename(2) needs
// write permission on the DIRECTORY, not on the file. Nothing in Centinela ever
// makes the file read-only, and 0644 is the mode every reader needs.
func TestSaveNormalisesAReadOnlyStateFile(t *testing.T) {
	stateRepo(t)
	if err := Save(New("alpha")); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(FilePath("alpha"), 0o400); err != nil {
		t.Fatal(err)
	}
	wf, err := Load("alpha")
	if err != nil {
		t.Fatal(err)
	}
	wf.CurrentStep = "code"
	if err := Save(wf); err != nil {
		t.Fatalf("a read-only state file must not block a save: %v", err)
	}
	info, err := os.Stat(FilePath("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("mode = %v, want 0644", info.Mode().Perm())
	}
}
