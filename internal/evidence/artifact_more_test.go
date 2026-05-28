package evidence

import (
	"os"
	"strings"
	"testing"
)

func TestRenderTemplateUnsupportedKindErrors(t *testing.T) {
	if _, _, err := RenderTemplate(ArtifactKind("bogus"), "alpha"); err == nil {
		t.Fatal("expected unsupported kind error")
	}
}

func TestParseKindHappyPath(t *testing.T) {
	for _, k := range KindsAllowed() {
		got, err := ParseKind(string(k))
		if err != nil || got != k {
			t.Fatalf("ParseKind(%s): got=%v err=%v", k, got, err)
		}
	}
}

func TestWriteArtifactGatekeeperHasStatusLine(t *testing.T) {
	chdirArtifactTemp(t)
	paths, err := WriteArtifact("alpha", KindGatekeeper, false)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := readPath(t, paths[0])
	if !strings.Contains(body, "**Status:** SAFE") {
		t.Fatalf("missing status line: %s", body)
	}
}

func TestWriteArtifactProductionReadinessHasStatusLine(t *testing.T) {
	chdirArtifactTemp(t)
	paths, err := WriteArtifact("alpha", KindProductionReadiness, false)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := readPath(t, paths[0])
	if !strings.Contains(body, "**Status:** PASS") {
		t.Fatalf("missing status line: %s", body)
	}
}

func TestWriteArtifactPropagatesRenderError(t *testing.T) {
	chdirArtifactTemp(t)
	if _, err := WriteArtifact("alpha", ArtifactKind("bogus"), false); err == nil {
		t.Fatal("expected RenderTemplate error to propagate")
	}
}

func TestWriteArtifactMkdirFailure(t *testing.T) {
	chdirArtifactTemp(t)
	// Place a regular file where `.workflow/` would live so MkdirAll fails.
	if err := os.WriteFile(".workflow", []byte("blocker"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteArtifact("alpha", KindEdgeCases, false); err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestWriteArtifactDocsJSONIsPretty(t *testing.T) {
	chdirArtifactTemp(t)
	paths, err := WriteArtifact("alpha", KindDocumentationSpecialist, false)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := readPath(t, paths[0])
	if !strings.Contains(body, "\n  \"feature\": \"alpha\"") {
		t.Fatalf("docs JSON not pretty-printed: %s", body)
	}
}

func readPath(t *testing.T, path string) (string, error) {
	t.Helper()
	b, err := os.ReadFile(path)
	return string(b), err
}
