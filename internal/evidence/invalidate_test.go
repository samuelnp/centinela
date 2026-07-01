package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

func setupWF(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidateRemovesBothAndIdempotent(t *testing.T) {
	setupWF(t)
	for _, ext := range []string{".json", ".md"} {
		p := filepath.Join(".workflow", "f-validation-specialist"+ext)
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := Invalidate("f", Role("validation-specialist"))
	if err != nil || !removed {
		t.Fatalf("Invalidate = %v %v", removed, err)
	}
	if _, err := os.Stat(".workflow/f-validation-specialist.md"); !os.IsNotExist(err) {
		t.Fatal("md not removed")
	}
	// Idempotent: a second call finds nothing, removes nothing, errors nothing.
	removed, err = Invalidate("f", Role("validation-specialist"))
	if err != nil || removed {
		t.Fatalf("idempotent = %v %v", removed, err)
	}
}

func TestInvalidateSafetyNeverTouchesSource(t *testing.T) {
	setupWF(t)
	if err := os.MkdirAll("internal/myfeature", 0o755); err != nil {
		t.Fatal(err)
	}
	src := "internal/myfeature/service.go"
	if err := os.WriteFile(src, []byte("package x"), 0o644); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(".workflow/f-qa-senior.json", []byte("x"), 0o644) //nolint:errcheck
	if _, err := Invalidate("f", Role("qa-senior")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("source must survive: %v", err)
	}
}

func TestInvalidateArtifactIdempotent(t *testing.T) {
	setupWF(t)
	os.WriteFile(".workflow/f-edge-cases.md", []byte("x"), 0o644) //nolint:errcheck
	removed, err := InvalidateArtifact("f", "edge-cases.md")
	if err != nil || !removed {
		t.Fatalf("artifact = %v %v", removed, err)
	}
	removed, err = InvalidateArtifact("f", "edge-cases.md")
	if err != nil || removed {
		t.Fatalf("idempotent artifact = %v %v", removed, err)
	}
}

func TestRemoveBothSurfacesRealError(t *testing.T) {
	setupWF(t)
	// A non-empty directory at the path makes os.Remove fail with a non-absence
	// error, exercising removeBoth's default arm.
	if err := os.MkdirAll(".workflow/f-gatekeeper.json/child", 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := removeBoth(".workflow/f-gatekeeper.json"); err == nil {
		t.Fatal("expected error removing non-empty dir")
	}
}
