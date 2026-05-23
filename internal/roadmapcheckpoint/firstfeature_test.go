package roadmapcheckpoint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func chdirTmp(t *testing.T) {
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
}

func bootstrap(features ...string) *roadmap.Roadmap {
	var fs []roadmap.Feature
	for _, n := range features {
		fs = append(fs, roadmap.Feature{Name: n})
	}
	return &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap", Features: fs}}}
}

func setStep(t *testing.T, feature, step string) {
	t.Helper()
	body := `{"feature":"` + feature + `","currentStep":"` + step + `","steps":{}}`
	if err := os.WriteFile(filepath.Join(".workflow", feature+".json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFirstIncompleteBootstrap_Cases(t *testing.T) {
	if name, ok := FirstIncompleteBootstrap(nil); ok || name != "" {
		t.Fatalf("nil roadmap -> (\"\", false), got (%q, %v)", name, ok)
	}

	chdirTmp(t)

	// No bootstrap phase.
	noBoot := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 1", Features: []roadmap.Feature{{Name: "x"}}}}}
	if name, ok := FirstIncompleteBootstrap(noBoot); ok || name != "" {
		t.Fatalf("no bootstrap -> (\"\", false), got (%q, %v)", name, ok)
	}

	// First non-done picked.
	if name, ok := FirstIncompleteBootstrap(bootstrap("alpha", "beta")); !ok || name != "alpha" {
		t.Fatalf("expected alpha, got (%q, %v)", name, ok)
	}

	// alpha done -> beta.
	setStep(t, "alpha", "done")
	if name, ok := FirstIncompleteBootstrap(bootstrap("alpha", "beta")); !ok || name != "beta" {
		t.Fatalf("expected beta, got (%q, %v)", name, ok)
	}

	// in-progress is non-done.
	setStep(t, "beta", "code")
	if name, ok := FirstIncompleteBootstrap(bootstrap("beta")); !ok || name != "beta" {
		t.Fatalf("in-progress beta is incomplete, got (%q, %v)", name, ok)
	}

	// all done -> none.
	setStep(t, "beta", "done")
	if name, ok := FirstIncompleteBootstrap(bootstrap("alpha", "beta")); ok || name != "" {
		t.Fatalf("all done -> (\"\", false), got (%q, %v)", name, ok)
	}
}
