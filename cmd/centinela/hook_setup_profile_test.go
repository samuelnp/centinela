package main

import (
	"os"
	"strings"
	"testing"
)

// guidedTree lays a project with PROJECT.md and a bootstrap roadmap json but
// none of the graded artifacts, under no centinela.toml (so it inherits guided).
func guidedTree(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("x"), 0644) //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json",
		[]byte(`{"phases":[{"name":"Phase 0: Bootstrap","features":[{"name":"setup"}]}]}`), 0644) //nolint:errcheck
}

// TestHookSetupGuidedAdvisesAndReachesCheckpoint: the guided cascade names every
// missing optional artifact once and still gets to the roadmap checkpoint.
func TestHookSetupGuidedAdvisesAndReachesCheckpoint(t *testing.T) {
	guidedTree(t)
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "Advisory: optional setup artifacts missing") {
			t.Fatalf("expected one consolidated advisory, got %q", out)
		}
		for _, want := range []string{"ROADMAP.md", "roadmap-analysis", "roadmap-quality",
			"production-readiness-prompt.md"} {
			if !strings.Contains(out, want) {
				t.Errorf("advisory must name %s, got %q", want, out)
			}
		}
		if strings.Contains(out, "CENTINELA DIRECTIVE: roadmap required") {
			t.Errorf("guided must not halt on ROADMAP.md, got %q", out)
		}
		if !strings.Contains(out, "roadmap checkpoint") {
			t.Errorf("guided must still reach the checkpoint, got %q", out)
		}
	})
}

// TestHookSetupGuidedStillRequiresRoadmapJSON: PROJECT.md and a parseable
// roadmap json are required in EVERY profile.
func TestHookSetupGuidedStillRequiresRoadmapJSON(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("x"), 0644) //nolint:errcheck
	withStdin(t, "{}", func() {
		out := captureStdout(t, func() { _ = runHookSetup(nil, nil) })
		if !strings.Contains(out, "roadmap json required") {
			t.Fatalf("guided must still demand the roadmap json, got %q", out)
		}
	})
}

// TestSetupRequiresGrading covers the knob both directions, including the
// fail-safe: an unparseable centinela.toml keeps the heavy cascade.
func TestSetupRequiresGrading(t *testing.T) {
	cases := []struct {
		name string
		toml string
		want bool
	}{
		{"no config inherits guided", "", false},
		{"explicit strict", "[workflow]\nenforcement_profile = \"strict\"\n", true},
		{"explicit guided", "[workflow]\nenforcement_profile = \"guided\"\n", false},
		{"explicit outcome", "[workflow]\nenforcement_profile = \"outcome\"\n", false},
		{"capable driver", "[orchestration]\ndriver_model = \"sonnet\"\n", false},
		{"limited driver", "[orchestration]\ndriver_model = \"haiku\"\n", true},
		{"unparseable config fails safe to grading", "this is not toml =", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("CENTINELA_MODEL", "")
			if tc.toml != "" {
				os.WriteFile("centinela.toml", []byte(tc.toml), 0644) //nolint:errcheck
			}
			if got := setupRequiresGrading(); got != tc.want {
				t.Fatalf("setupRequiresGrading() = %v, want %v", got, tc.want)
			}
		})
	}
}
