package filelock

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func lockIn(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "nested", "state.lock")
}

// The whole point: a second holder must not get in while the first holds it.
func TestAcquireExcludesASecondHolder(t *testing.T) {
	path := lockIn(t)
	release, err := Acquire(path, DefaultTimeout, DefaultPollInterval, "")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if _, err := Acquire(path, 150*time.Millisecond, 10*time.Millisecond, ""); err == nil {
		t.Fatal("a second acquire must not succeed while the lock is held")
	}
	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Fatalf("it gave up after %v — it did not actually wait", elapsed)
	}
	release()
	release2, err := Acquire(path, DefaultTimeout, DefaultPollInterval, "")
	if err != nil {
		t.Fatalf("the lock must be reusable after release: %v", err)
	}
	release2()
}

// The parent directory is created on demand, so a first-ever mutation works.
func TestAcquireCreatesTheParentDirectory(t *testing.T) {
	release, err := Acquire(lockIn(t), DefaultTimeout, DefaultPollInterval, "")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	release()
}

// The timeout names the file and carries the caller's hint, so an operator can
// tell "someone else is mid-write" from a crash.
func TestAcquireTimeoutNamesTheFileAndTheHint(t *testing.T) {
	path := lockIn(t)
	release, err := Acquire(path, DefaultTimeout, DefaultPollInterval, "")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	_, err = Acquire(path, 20*time.Millisecond, 5*time.Millisecond, "retry the mutation")
	if err == nil {
		t.Fatal("want a timeout")
	}
	for _, want := range []string{"lock busy", path, "another process is writing", "retry the mutation"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("%q missing from %q", want, err)
		}
	}
	if _, err := Acquire(path, 20*time.Millisecond, 5*time.Millisecond, ""); err == nil ||
		strings.Contains(err.Error(), " — ") {
		t.Fatalf("an empty hint must add no separator: %v", err)
	}
}
