package gitutil

import (
	"errors"
	"os/exec"
	"testing"
)

// realExitError returns a genuine *exec.ExitError (a non-zero process exit),
// the shape `git remote get-url origin` produces when origin is absent.
func realExitError() error {
	return exec.Command("sh", "-c", "exit 3").Run()
}

func stubGitRun(t *testing.T, out string, err error) {
	t.Helper()
	prev := gitRun
	gitRun = func(string, ...string) (string, error) { return out, err }
	t.Cleanup(func() { gitRun = prev })
}

// TestHasOriginRemotePresent: a successful non-empty URL means yes.
func TestHasOriginRemotePresent(t *testing.T) {
	stubGitRun(t, "git@github.com:o/r.git", nil)
	ok, err := HasOriginRemote(".")
	if err != nil || !ok {
		t.Fatalf("present: ok=%v err=%v", ok, err)
	}
}

// TestHasOriginRemoteAbsent: a non-zero git exit (ExitError) means no, not an error.
func TestHasOriginRemoteAbsent(t *testing.T) {
	stubGitRun(t, "", realExitError())
	ok, err := HasOriginRemote(".")
	if err != nil || ok {
		t.Fatalf("absent should be (false,nil): ok=%v err=%v", ok, err)
	}
}

// TestHasOriginRemoteEmptyURL: success but empty output means no.
func TestHasOriginRemoteEmptyURL(t *testing.T) {
	stubGitRun(t, "", nil)
	if ok, _ := HasOriginRemote("."); ok {
		t.Fatal("empty url should be false")
	}
}

// TestHasOriginRemoteRealError: a non-ExitError failure (e.g. git missing) propagates.
func TestHasOriginRemoteRealError(t *testing.T) {
	stubGitRun(t, "", errors.New("exec: git not found"))
	if _, err := HasOriginRemote("."); err == nil {
		t.Fatal("a real exec failure must propagate")
	}
}
