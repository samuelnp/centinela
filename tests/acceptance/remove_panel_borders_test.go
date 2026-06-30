package acceptance_test

// Acceptance: specs/remove-panel-borders.feature

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var bordersBinOnce sync.Once
var bordersBin string

func buildBordersBin(t *testing.T) string {
	t.Helper()
	bordersBinOnce.Do(func() {
		dir, _ := os.MkdirTemp("", "cent-borders-bin")
		bordersBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", bordersBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return bordersBin
}

// Scenario: a CLI command panel renders without a border.
func TestAccRoadmapPanelHasNoBorder(t *testing.T) {
	bin := buildBordersBin(t)
	out, code := runCent(t, bin, repoRoot(t), "roadmap")
	if code != 0 {
		t.Fatalf("roadmap exited %d: %s", code, out)
	}
	if strings.ContainsAny(out, "╭╮╰╯│") {
		t.Fatalf("roadmap panel should have no border box, got:\n%s", out)
	}
	if !strings.Contains(out, "PHASE OVERVIEW") || !strings.Contains(out, "🛡️👁️") {
		t.Errorf("roadmap panel lost its branded header")
	}
}
