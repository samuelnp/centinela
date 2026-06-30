package selfupdate

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReplaceBinaryAtomic(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "bin")
	if err := os.WriteFile(target, []byte("old"), 0o711); err != nil {
		t.Fatal(err)
	}
	if err := replaceBinary(target, []byte("new")); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new" {
		t.Fatalf("not replaced: %q", got)
	}
	info, _ := os.Stat(target)
	if info.Mode().Perm() != 0o711 {
		t.Fatalf("mode bits not preserved: %v", info.Mode())
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("leftover temp files: %d", len(entries))
	}
}

func TestReplaceBinaryMissingTarget(t *testing.T) {
	err := replaceBinary(filepath.Join(t.TempDir(), "nope"), []byte("x"))
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindReplace {
		t.Fatalf("want replace error, got %v", err)
	}
}

func TestTargetPathResolves(t *testing.T) {
	p, err := targetPath()
	if err != nil || p == "" {
		t.Fatalf("targetPath: %v %q", err, p)
	}
}

func TestCacheRoundTripAndStale(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	now := time.Unix(1_700_000_000, 0)
	u := &Updater{TTL: time.Hour, Now: func() time.Time { return now }}
	u.writeCache("v1.2.3")
	if tag, ok := u.readCache(); !ok || tag != "v1.2.3" {
		t.Fatalf("round trip: %q ok=%v", tag, ok)
	}
	u.Now = func() time.Time { return now.Add(2 * time.Hour) }
	if _, ok := u.readCache(); ok {
		t.Fatal("stale cache should miss")
	}
}

func TestCacheMissingAndTTLDefault(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	u := &Updater{Now: time.Now}
	if _, ok := u.readCache(); ok {
		t.Fatal("missing cache should miss")
	}
	if u.ttl() != defaultTTL {
		t.Fatalf("default ttl = %v", u.ttl())
	}
}
