package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
)

const driftJSON = `{"phases":[{"name":"P","features":[{"name":"a","description":"d"}]}]}`

// seedDrift chdirs into a temp repo with roadmap.json and (optionally) a
// ROADMAP.md, returning the in-sync generator output for the seeded roadmap.
func seedDrift(t *testing.T, roadmapMD *string) []byte {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".workflow", "roadmap.json"), []byte(driftJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	r, _ := roadmap.Load()
	want := roadmap.RenderMarkdown(r)
	if roadmapMD != nil {
		if err := os.WriteFile(filepath.Join(dir, "ROADMAP.md"), []byte(*roadmapMD), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return want
}

func driftCfg(sev string) *config.Config {
	return &config.Config{Gates: config.GatesConfig{RoadmapDrift: config.RoadmapDriftConfig{Enabled: true, Severity: sev}}}
}

func TestCheckRoadmapDriftInSync(t *testing.T) {
	want := seedDrift(t, nil)
	s := string(want)
	seedDrift(t, &s)
	r := checkRoadmapDrift(driftCfg("fail"), nil)
	if r.Status != Pass {
		t.Fatalf("in-sync must Pass, got %v: %s", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "in sync") {
		t.Fatalf("message: %q", r.Message)
	}
}

func TestCheckRoadmapDriftFail(t *testing.T) {
	bad := "# Roadmap\n\nhand-edited\n"
	seedDrift(t, &bad)
	r := checkRoadmapDrift(driftCfg("fail"), nil)
	if r.Status != Fail {
		t.Fatalf("drift under fail must Fail, got %v", r.Status)
	}
	if !strings.Contains(r.Message, "line") || !strings.Contains(r.Message, "roadmap generate") {
		t.Fatalf("message must name the line and remediation: %q", r.Message)
	}
}

func TestCheckRoadmapDriftWarn(t *testing.T) {
	bad := "# Roadmap\n\nhand-edited\n"
	seedDrift(t, &bad)
	r := checkRoadmapDrift(driftCfg("warn"), nil)
	if r.Status != Warn {
		t.Fatalf("drift under warn must Warn, got %v", r.Status)
	}
}

func TestCheckRoadmapDriftMissingFile(t *testing.T) {
	seedDrift(t, nil)
	r := checkRoadmapDrift(driftCfg("fail"), nil)
	if r.Status != Fail || !strings.Contains(r.Message, "missing") {
		t.Fatalf("missing ROADMAP.md must Fail with 'missing', got %v %q", r.Status, r.Message)
	}
	rw := checkRoadmapDrift(driftCfg("warn"), nil)
	if rw.Status != Warn || !strings.Contains(rw.Message, "missing") {
		t.Fatalf("missing under warn must Warn, got %v %q", rw.Status, rw.Message)
	}
}
