package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Acceptance: specs/capability-calibration.feature
//
// Shared helpers. calRepo builds a temp project whose centinela.toml declares the
// three spec capability classes, optionally seeding the telemetry log; calBin is a
// once-built binary shared across every calibration scenario.

const calToml = `[orchestration.capabilities]
"claude-opus-4-7" = "frontier"
"claude-sonnet-4-6" = "capable"
"claude-haiku-4-5" = "limited"
`

var (
	calBinOnce sync.Once
	calBin     string
	calBinErr  string
)

// calRepo creates a temp repo with the capability config. A nil lines leaves the
// log absent (missing-log scenario); an empty slice writes an empty file; else it
// seeds events.jsonl with each line.
func calRepo(t *testing.T, lines []string) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	if err := os.WriteFile(filepath.Join(dir, "centinela.toml"), []byte(calToml), 0o644); err != nil {
		t.Fatal(err)
	}
	if lines == nil {
		return dir
	}
	td := filepath.Join(dir, ".workflow", "telemetry")
	if err := os.MkdirAll(td, 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Join(lines, "\n")
	if len(lines) > 0 {
		body += "\n"
	}
	if err := os.WriteFile(filepath.Join(td, "events.jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// buildCalBin compiles centinela once into a persistent temp dir.
func buildCalBin(t *testing.T) string {
	t.Helper()
	calBinOnce.Do(func() {
		bin, err := os.MkdirTemp("", "cent-cal-bin")
		if err != nil {
			calBinErr = err.Error()
			return
		}
		calBin = filepath.Join(bin, "centinela")
		c := buildCalCmd(t, calBin)
		if out, err := c.CombinedOutput(); err != nil {
			calBinErr = err.Error() + "\n" + string(out)
			calBin = ""
		}
	})
	if calBin == "" {
		t.Fatalf("build centinela: %s", calBinErr)
	}
	return calBin
}

// runCal runs `calibrate [args...]` in a seeded repo, returning output + code.
func runCal(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildCalBin(t), dir, append([]string{"calibrate"}, args...)...)
}

// adv/gf/vr build JSONL lines for a model's events (helpers keep scenarios terse).
func adv(model string) string { return calLine("step-advanced", model) }
func gf(model string) string  { return calLine("gate-failure", model) }
func vr(model string) string  { return calLine("verify-rejection", model) }
