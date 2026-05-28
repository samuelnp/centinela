package evidence

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdirArtifactTemp(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
}

func TestRenderTemplateAllKinds(t *testing.T) {
	for _, kind := range KindsAllowed() {
		paths, bodies, err := RenderTemplate(kind, "alpha")
		if err != nil {
			t.Fatalf("%s render: %v", kind, err)
		}
		if len(paths) == 0 || len(paths) != len(bodies) {
			t.Fatalf("%s: paths/bodies mismatch %d/%d", kind, len(paths), len(bodies))
		}
		for _, p := range paths {
			if !strings.HasPrefix(p, ".workflow/alpha-") {
				t.Fatalf("%s: path not scoped: %s", kind, p)
			}
		}
	}
}

func TestParseKindRejectsUnknown(t *testing.T) {
	if _, err := ParseKind("bogus"); err == nil ||
		!strings.Contains(err.Error(), "edge-cases") {
		t.Fatalf("expected allowed-kinds list in err, got %v", err)
	}
}

func TestWriteArtifactRefusesOverwriteWithoutForce(t *testing.T) {
	chdirArtifactTemp(t)
	if _, err := WriteArtifact("alpha", KindEdgeCases, false); err != nil {
		t.Fatal(err)
	}
	_, err := WriteArtifact("alpha", KindEdgeCases, false)
	if !errors.Is(err, ErrArtifactExists) {
		t.Fatalf("expected ErrArtifactExists, got %v", err)
	}
}

func TestWriteArtifactForceOverwrites(t *testing.T) {
	chdirArtifactTemp(t)
	if _, err := WriteArtifact("alpha", KindGatekeeper, false); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(".workflow", "alpha-gatekeeper.md")
	_ = os.WriteFile(path, []byte("DIRTY"), 0o644)
	if _, err := WriteArtifact("alpha", KindGatekeeper, true); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "DIRTY") {
		t.Fatalf("force did not overwrite: %s", data)
	}
}

func TestWriteArtifactDocsPair(t *testing.T) {
	chdirArtifactTemp(t)
	paths, err := WriteArtifact("alpha", KindDocumentationSpecialist, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %v", paths)
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("path missing: %s err=%v", p, err)
		}
	}
}
