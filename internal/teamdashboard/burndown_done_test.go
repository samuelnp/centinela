package teamdashboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestBurndown_DoneCountsFromWorkflowStatus seeds a done workflow file on disk
// so roadmap.FeatureStatus resolves "done", exercising the PhaseStatus.Done++
// branch and the Summary() done tally.
func TestBurndown_DoneCountsFromWorkflowStatus(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	js := `{"feature":"shipped","currentStep":"done","steps":{}}`
	if err := os.WriteFile(filepath.Join(".workflow", "shipped.json"), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{phase("Q1", "shipped", "pending")}}
	b := burndown(r)
	if b.Done != 1 || b.Planned != 1 || b.Total != 2 {
		t.Fatalf("summary: %+v want Done=1 Planned=1 Total=2", b)
	}
	if len(b.Phases) != 1 || b.Phases[0].Done != 1 || b.Phases[0].Total != 2 {
		t.Fatalf("phase done/total: %+v", b.Phases)
	}
}
