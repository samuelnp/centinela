package verify

import (
	"time"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// fakeRunner is a CommandRunner that returns a scripted outcome per command,
// falling back to def for unscripted commands. It records the calls it saw.
type fakeRunner struct {
	byCmd map[string]RunOutcome
	def   RunOutcome
	calls []string
}

func (f *fakeRunner) Run(_ string, command string, _ time.Duration) RunOutcome {
	f.calls = append(f.calls, command)
	if out, ok := f.byCmd[command]; ok {
		return out
	}
	return f.def
}

// fakeLoad builds an EvidenceLoader returning ev for the qa-senior role.
func fakeLoad(ev *evidence.RoleEvidence, err error) EvidenceLoader {
	return func(_ string, _ orchestration.Role) (*evidence.RoleEvidence, error) {
		return ev, err
	}
}

func cov(v float64) *float64 { return &v }
