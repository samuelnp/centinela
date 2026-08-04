package filelock

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// An unopenable path is a real error, never a silently unlocked success.
func TestAcquireFailsClosedOnAnUnopenablePath(t *testing.T) {
	file := filepath.Join(t.TempDir(), "regular")
	release, err := Acquire(file, DefaultTimeout, DefaultPollInterval, "")
	if err != nil {
		t.Fatal(err)
	}
	release()
	// The lock file now exists as a file; using it as a DIRECTORY must fail.
	if _, err := Acquire(filepath.Join(file, "child.lock"), DefaultTimeout, DefaultPollInterval, ""); err == nil {
		t.Fatal("want an error when the parent cannot be created")
	}
}

// Serialization under real contention: every worker's increment must survive.
func TestAcquireSerializesConcurrentWorkers(t *testing.T) {
	path := lockIn(t)
	const n = 12
	var wg sync.WaitGroup
	var counter int
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := Acquire(path, 5*time.Second, time.Millisecond, "")
			if err != nil {
				t.Errorf("acquire: %v", err)
				return
			}
			defer release()
			counter++ // guarded solely by the file lock
		}()
	}
	wg.Wait()
	if counter != n {
		t.Fatalf("counter = %d, want %d", counter, n)
	}
}

// A DIRECTORY where the lock file belongs cannot be opened for writing, and
// must surface as an error rather than an unlocked "success".
func TestAcquireFailsClosedWhenTheLockPathIsADirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.lock")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Acquire(path, DefaultTimeout, DefaultPollInterval, "")
	if err == nil || !strings.Contains(err.Error(), "lock open") {
		t.Fatalf("want a lock-open error, got %v", err)
	}
}

// A real lock failure must never be mistaken for "busy": busy means retry,
// an error means stop. A closed descriptor is the cheapest genuine failure.
func TestTryLockExclusiveDistinguishesFailureFromBusy(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "closed-*.lock")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	locked, err := tryLockExclusive(f)
	if locked || err == nil {
		t.Fatalf("a closed descriptor must report an error, got locked=%v err=%v", locked, err)
	}
	if err := unlockFile(f); err == nil {
		t.Fatal("unlocking a closed descriptor must also surface an error")
	}
}
