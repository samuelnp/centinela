package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/roadmap"
)

const rdsIntegrationJSON = `{"intro":"hi","phases":[
 {"name":"P","note":"why","features":[{"name":"a","description":"d","fixes":"f"}]},
 {"name":"Backlog","features":[{"name":"b","summary":"s","deferredAt":"t"}]}]}`

func driftResultIn(results []gates.Result) gates.Result {
	for _, r := range results {
		if r.Name == "roadmap_drift" {
			return r
		}
	}
	return gates.Result{}
}

// Generate → gate round-trip: in sync Passes; a hand-edit Fails; regenerate Passes.
func TestRoadmapDocSyncRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".workflow/roadmap.json", []byte(rdsIntegrationJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Gates: config.GatesConfig{
		RoadmapDrift: config.RoadmapDriftConfig{Enabled: true, Severity: "fail"}}}

	// Generate ROADMAP.md from roadmap.json.
	r, err := roadmap.Load()
	if err != nil {
		t.Fatal(err)
	}
	md := roadmap.RenderMarkdown(r)
	if err := os.WriteFile("ROADMAP.md", md, 0o644); err != nil {
		t.Fatal(err)
	}

	// In sync → Pass.
	if got := driftResultIn(gates.RunAll(cfg)); got.Status != gates.Pass {
		t.Fatalf("in-sync want Pass, got %v: %s", got.Status, got.Message)
	}

	// Mutate → Fail with line + remediation.
	if err := os.WriteFile("ROADMAP.md", []byte("# Roadmap\n\ntampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := driftResultIn(gates.RunAll(cfg))
	if got.Status != gates.Fail {
		t.Fatalf("drift want Fail, got %v", got.Status)
	}
	if !strings.Contains(got.Message, "line") || !strings.Contains(got.Message, "roadmap generate") {
		t.Fatalf("drift message lacks line/remediation: %q", got.Message)
	}

	// Regenerate → Pass again.
	if err := os.WriteFile("ROADMAP.md", md, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := driftResultIn(gates.RunAll(cfg)); got.Status != gates.Pass {
		t.Fatalf("after regen want Pass, got %v", got.Status)
	}

	// Disabled gate is absent from results.
	cfg.Gates.RoadmapDrift.Enabled = false
	if driftResultIn(gates.RunAll(cfg)).Name != "" {
		t.Fatal("disabled gate must not appear in results")
	}
}
