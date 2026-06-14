package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// buildDoctor compiles the centinela binary into a persistent temp dir.
func buildDoctor(t *testing.T) string {
	t.Helper()
	tmp, err := os.MkdirTemp("", "cent-doctor-int")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(tmp, "centinela")
	wd, _ := os.Getwd() // tests/integration
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	c := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	c.Dir = root
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func runDoc(t *testing.T, bin, dir string, args ...string) (string, int) {
	t.Helper()
	c := exec.Command(bin, append([]string{"doctor"}, args...)...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	return string(out), code
}

// seedDriftedRepo builds a git repo with: missing hooks, drifted ROADMAP.md, a
// phase-name glyph, and an orphaned *.json.tmp — every safe-repair path at once.
func seedDriftedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	for _, a := range [][]string{
		{"init", "-q", "-b", "main"}, {"config", "user.email", "i@c.dev"}, {"config", "user.name", "I"},
	} {
		c := exec.Command("git", a...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", a, err, out)
		}
	}
	_ = os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755)
	_ = os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte("{}"), 0o644)
	rm := &roadmap.Roadmap{Phases: []roadmap.Phase{{
		Name: "✅ Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "f", Description: "d"}},
	}}}
	cur, _ := os.Getwd()
	_ = os.Chdir(dir)
	_ = roadmap.Save(rm)
	_ = os.Chdir(cur)
	_ = os.WriteFile(filepath.Join(dir, "ROADMAP.md"), []byte("drifted\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, ".workflow", "f-qa-senior.json.tmp"), []byte("{}"), 0o644)
	return dir
}
