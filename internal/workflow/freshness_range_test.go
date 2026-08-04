package workflow

import (
	"errors"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/treestate"
)

// AC9's headline: a deferral commits itself after the stamp and the verdict
// must survive it.
func TestVerificationSurvivesARoadmapStateOnlyCommit(t *testing.T) {
	stamped := gitStubRange("abc123", "", "")
	seedGate(t, ValidateContractAdversarial, freshReport(t, stamped))
	moved := gitStubRange("def456", "", ".workflow/roadmap.json\nROADMAP.md\n")
	if err := VerificationFresh("f", ".", moved); err != nil {
		t.Fatalf("a roadmap-state-only commit must not stale the verification: %v", err)
	}
}

// The direction that keeps the exemption from becoming a hole (R1).
func TestVerificationStalesOnASourceChangeBesideRoadmapState(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, freshReport(t, gitStubRange("abc123", "", "")))
	moved := gitStubRange("def456", "", ".workflow/roadmap.json\nROADMAP.md\ninternal/example/thing.go\n")
	err := VerificationFresh("f", ".", moved)
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("want stale, got %v", err)
	}
	for _, want := range []string{"abc123", "def456"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the message must name %s: %v", want, err)
		}
	}
}

func TestRevisionRangeStateOnlyDecisionTable(t *testing.T) {
	cases := []struct {
		name, out string
		err       error
		want      bool
	}{
		{"state only", ".workflow/roadmap.json\nROADMAP.md\n", nil, true},
		{"empty range", "", nil, true},
		{"blank lines only", "\n  \n", nil, true},
		{"mixed", "ROADMAP.md\ninternal/x.go\n", nil, false},
		{"source only", "internal/x.go\n", nil, false},
		{"git error fails closed", ".workflow/roadmap.json\n", errors.New("bad revision"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := treestate.Runner(func(string, ...string) (string, error) { return tc.out, tc.err })
			if got := revisionRangeIsRoadmapStateOnly(".", "abc123", run); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// An unstamped revision can never be exempt — there is no range to read.
func TestRevisionRangeStateOnlyRejectsAnEmptyRecordedRevision(t *testing.T) {
	run := treestate.Runner(func(string, ...string) (string, error) {
		t.Fatal("git must not be consulted without a recorded revision")
		return "", nil
	})
	for _, rev := range []string{"", "   "} {
		if revisionRangeIsRoadmapStateOnly(".", rev, run) {
			t.Fatalf("recorded revision %q must not be exempt", rev)
		}
	}
}

// The uncommitted half of the same exemption (disable_auto_commit).
func TestVerificationSurvivesAnUncommittedRegeneratedMarkdown(t *testing.T) {
	stamped := gitStubRange("abc123", "", "")
	seedGate(t, ValidateContractAdversarial, freshReport(t, stamped))
	dirty := gitStubRange("abc123", " M ROADMAP.md\n M .workflow/roadmap.json\n", "")
	if err := VerificationFresh("f", ".", dirty); err != nil {
		t.Fatalf("an uncommitted regenerated ROADMAP.md must not stale: %v", err)
	}
}
