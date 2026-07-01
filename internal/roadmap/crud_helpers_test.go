package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

// crudBody is a fixed multi-phase roadmap used across the add/remove/promote
// CRUD tests: a schedulable Phase 1/Phase 2, a solo phase, and a Backlog.
const crudBody = `{"phases":[` +
	`{"name":"Phase 1: Foundations","features":[{"name":"auth-service"},` +
	`{"name":"checkout-ui","dependsOn":["auth-service"]}]},` +
	`{"name":"Phase 2: Growth","features":[{"name":"billing-api"}]},` +
	`{"name":"Phase 3: Solo","features":[{"name":"lonely-feature"}]},` +
	`{"name":"Backlog","features":[{"name":"legacy-finding","summary":"s"}]}]}`

// crudWrite writes body to a standalone roadmap.json in a temp dir and returns
// the path. Used for Add/Remove tests that only touch the given path.
func crudWrite(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "roadmap.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// crudChdir writes body to .workflow/roadmap.json inside a fresh temp dir and
// chdirs into it (restoring cwd on cleanup), for tests exercising CWD-relative
// readers (FeatureStatus, artifact files).
func crudChdir(t *testing.T, body string) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(RoadmapFile, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// crudBytes returns the current on-disk bytes at path, failing the test on error.
func crudBytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}
