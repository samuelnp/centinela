package roadmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRemoveFeatureEntries_PrunesMatchedLeavesRest drops only the named entries.
func TestRemoveFeatureEntries_PrunesMatchedLeavesRest(t *testing.T) {
	p := filepath.Join(t.TempDir(), "analysis.json")
	body := `{"role":"senior-product-manager","features":[` +
		`{"name":"billing-api","summary":"s"},{"name":"reporting","summary":"s"},` +
		`{"name":"auth-service","summary":"keep"}]}`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := removeFeatureEntries(p, map[string]bool{"billing-api": true, "reporting": true}); err != nil {
		t.Fatalf("removeFeatureEntries: %v", err)
	}
	got := string(crudBytes(t, p))
	if strings.Contains(got, "billing-api") || strings.Contains(got, "reporting") {
		t.Fatalf("matched entries must be pruned: %s", got)
	}
	if !strings.Contains(got, "auth-service") || !strings.Contains(got, `"role"`) {
		t.Fatalf("untouched entry and top-level fields must survive: %s", got)
	}
}

// TestRemoveFeatureEntries_MissingFileNoop: an absent artifact file is a no-op.
func TestRemoveFeatureEntries_MissingFileNoop(t *testing.T) {
	p := filepath.Join(t.TempDir(), "absent.json")
	if err := removeFeatureEntries(p, map[string]bool{"x": true}); err != nil {
		t.Fatalf("missing file must be a no-op, got %v", err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("no-op must not create the file")
	}
}

// TestRemoveFeatureEntries_InvalidJSON surfaces an error, no crash.
func TestRemoveFeatureEntries_InvalidJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(p, []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	wantErr(t, removeFeatureEntries(p, map[string]bool{"x": true}), "invalid artifact json")
}
