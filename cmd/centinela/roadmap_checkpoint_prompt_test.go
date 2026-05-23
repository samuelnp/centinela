package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// chdirIntoTemp moves into a fresh temp dir and restores cwd on cleanup.
func chdirIntoTemp(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	return d
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// roadmapJSON with a single Phase 0 bootstrap phase containing the named features.
func roadmapJSON(features ...string) string {
	var b strings.Builder
	b.WriteString(`{"phases":[{"name":"Phase 0: Bootstrap","features":[`)
	for i, f := range features {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"name":"` + f + `"}`)
	}
	b.WriteString(`]}]}`)
	return b.String()
}

// layRoadmapArtifacts writes the full set of roadmap-defining artifacts plus
// the production-readiness prompt, so runHookSetup reaches the checkpoint branch.
func layRoadmapArtifacts(t *testing.T, roadmapBody string) {
	t.Helper()
	writeFile(t, "PROJECT.md", "x")
	writeFile(t, "ROADMAP.md", "x")
	writeFile(t, ".workflow/roadmap.json", roadmapBody)
	writeFile(t, ".workflow/roadmap-analysis.md", "x")
	writeFile(t, ".workflow/roadmap-analysis.json", "{}")
	writeFile(t, ".workflow/roadmap-quality.md", "x")
	writeFile(t, ".workflow/roadmap-quality.json", "{}")
	writeFile(t, "docs/architecture/production-readiness-prompt.md", "x")
}

func runSetup(t *testing.T) string {
	t.Helper()
	var out string
	withStdin(t, "{}", func() {
		out = captureStdout(t, func() {
			if err := runHookSetup(nil, nil); err != nil {
				t.Fatalf("runHookSetup returned error: %v", err)
			}
		})
	})
	return out
}

const ckptDirective = "CENTINELA DIRECTIVE: roadmap checkpoint"

func assertContains(t *testing.T, out, want string) {
	t.Helper()
	if !strings.Contains(out, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, out)
	}
}

func assertNotContains(t *testing.T, out, notWant string) {
	t.Helper()
	if strings.Contains(out, notWant) {
		t.Fatalf("expected output NOT to contain %q, got:\n%s", notWant, out)
	}
}
