package golist

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// chdirTemp creates a temp dir, chdirs into it, and reverts on cleanup.
func chdirTemp(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	return d
}

// writeT writes body to dir/rel, creating parents.
func writeT(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// stubGo writes a fake `go` shell binary running body and prepends it to PATH so
// runGo invokes it instead of the real toolchain (deterministic edge coverage).
func stubGo(t *testing.T, body string) {
	t.Helper()
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "go"), []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", d+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestRunGo_EmptyStderrWrapsRawError(t *testing.T) {
	// Non-zero exit with NO stderr -> the firstStderrLine=="" fallback wraps the
	// raw exec error so callers never treat a failure as empty success.
	stubGo(t, "exit 3\n")
	_, err := Packages()
	if err == nil || !strings.Contains(err.Error(), "go list -json ./...") {
		t.Fatalf("empty-stderr failure must wrap the raw error: %v", err)
	}
}

func TestPackages_DecodeErrorSurfaced(t *testing.T) {
	// Exit 0 but emit non-JSON on stdout -> the streamed decode fails and the
	// error is surfaced rather than returning a silently-truncated package list.
	stubGo(t, "printf 'this is not json'\nexit 0\n")
	_, err := Packages()
	if err == nil || !strings.Contains(err.Error(), "decoding go list output") {
		t.Fatalf("malformed go list output must surface a decode error: %v", err)
	}
}
