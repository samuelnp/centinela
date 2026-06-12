package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// NewWithOrder pins the profile and sets the orchestration mode only for strict:
// strict requires subagent evidence (StrictOrchestrationMode); guided/outcome
// leave it empty so validateOrchestration early-returns (evidence optional).
func TestNewWithOrder_OrchestrationModePerProfile(t *testing.T) {
	cases := []struct {
		profile  string
		wantMode string
	}{
		{config.ProfileStrict, StrictOrchestrationMode},
		{config.ProfileGuided, ""},
		{config.ProfileOutcome, ""},
		{"", StrictOrchestrationMode},      // empty normalizes to strict
		{"bogus", StrictOrchestrationMode}, // unknown normalizes to strict
	}
	for _, c := range cases {
		wf := NewWithOrder("f", DefaultStepOrder, c.profile)
		if wf.OrchestrationMode != c.wantMode {
			t.Fatalf("profile %q: OrchestrationMode = %q, want %q", c.profile, wf.OrchestrationMode, c.wantMode)
		}
		wantPinned := config.NormalizeEnforcementProfile(c.profile)
		if wf.EnforcementProfile != wantPinned {
			t.Fatalf("profile %q: pinned = %q, want %q", c.profile, wf.EnforcementProfile, wantPinned)
		}
	}
}

func TestNew_DefaultsToStrict(t *testing.T) {
	wf := New("f")
	if wf.EnforcementProfile != config.ProfileStrict || wf.OrchestrationMode != StrictOrchestrationMode {
		t.Fatalf("New must pin strict + strict-subagents, got profile=%q mode=%q",
			wf.EnforcementProfile, wf.OrchestrationMode)
	}
}
