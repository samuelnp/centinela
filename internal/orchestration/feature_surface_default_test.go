package orchestration

import (
	"os"
	"testing"
)

func TestIsUserFacingFeatureDefaultsToInternalWhenNoSurface(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	// A brief that declares no surface line at all.
	if err := os.WriteFile("docs/features/none.md", []byte("# none\njust a title\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if IsUserFacingFeature("none") {
		t.Fatal("a brief with no surface line must default to internal (false)")
	}
}

func TestIsUserFacingFeatureDefaultsToInternalWhenBriefMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	if IsUserFacingFeature("absent") {
		t.Fatal("a missing brief must default to internal (false)")
	}
}
