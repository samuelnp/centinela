package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// sectionBody returns the text of the report section whose title starts with
// title, up to the next blank-line-separated section. Sections are joined by
// "\n\n" in the renderer, so a split on "\n\n" isolates each block.
func sectionBody(out, title string) string {
	for _, block := range strings.Split(out, "\n\n") {
		if strings.HasPrefix(strings.TrimSpace(block), title) {
			return block
		}
	}
	return ""
}

var (
	insightsBinOnce sync.Once
	insightsBin     string
	insightsBinErr  string
)

// insightsRepo builds a temp project dir; if lines is non-nil it seeds
// .workflow/telemetry/events.jsonl with each line. A nil lines leaves the log
// absent (missing-log scenario); an empty slice writes an empty file.
func insightsRepo(t *testing.T, lines []string) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	if lines == nil {
		return dir
	}
	td := filepath.Join(dir, ".workflow", "telemetry")
	if err := os.MkdirAll(td, 0o755); err != nil {
		t.Fatal(err)
	}
	body := ""
	for _, l := range lines {
		body += l + "\n"
	}
	if err := os.WriteFile(filepath.Join(td, "events.jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// buildInsightsBin compiles centinela once into a persistent temp dir, shared
// across every insights scenario (mirrors buildDoctorBin).
func buildInsightsBin(t *testing.T) string {
	t.Helper()
	insightsBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-insights-bin")
		if err != nil {
			insightsBinErr = err.Error()
			return
		}
		bin := filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			insightsBinErr = err.Error() + "\n" + string(out)
			return
		}
		insightsBin = bin
	})
	if insightsBin == "" {
		t.Fatalf("build centinela: %s", insightsBinErr)
	}
	return insightsBin
}

// runInsights runs `insights [args...]` in a seeded repo, returning output+code.
func runInsights(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildInsightsBin(t), dir, append([]string{"insights"}, args...)...)
}
