package planadvisor

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/insights"
)

func TestWorstGateAndTopFailureCountEmpty(t *testing.T) {
	b := bundle{}
	if worstGate(b) != "" {
		t.Fatalf("empty worstGate = %q, want \"\"", worstGate(b))
	}
	if topFailureCount(b) != 0 {
		t.Fatalf("empty topFailureCount = %d, want 0", topFailureCount(b))
	}
}

func TestWorstGateAndTopFailureCountPopulated(t *testing.T) {
	b := bundle{Failures: []insights.Count{{Key: "g1-file-size", Count: 5}, {Key: "coverage", Count: 2}}}
	if worstGate(b) != "g1-file-size" {
		t.Fatalf("worstGate = %q", worstGate(b))
	}
	if topFailureCount(b) != 5 {
		t.Fatalf("topFailureCount = %d, want 5", topFailureCount(b))
	}
}

// hasFailureQuestion reports whether the selected questions include the
// gate-failure pre-warning naming the worst gate.
func hasFailureQuestion(qs []question, gate string) bool {
	for _, q := range qs {
		if strings.Contains(q.Text, "recurring gate failures") && strings.Contains(q.Text, gate) {
			return true
		}
	}
	return false
}

func TestSelectQuestionsIncludesPreWarningAtThreshold(t *testing.T) {
	b := bundle{Feature: "f", Failures: []insights.Count{{Key: "g1-file-size", Count: 2}}}
	qs := selectQuestions(b, 99, "always")
	if !hasFailureQuestion(qs, "g1-file-size") {
		t.Fatalf("expected pre-warning question naming g1-file-size, got %+v", qs)
	}
}

func TestSelectQuestionsOmitsPreWarningBelowThreshold(t *testing.T) {
	b := bundle{Feature: "f", Failures: []insights.Count{{Key: "g1-file-size", Count: 1}}}
	qs := selectQuestions(b, 99, "always")
	if hasFailureQuestion(qs, "g1-file-size") {
		t.Fatalf("count 1 must not produce a pre-warning question, got %+v", qs)
	}
	// And no failures at all stays silent too.
	if hasFailureQuestion(selectQuestions(bundle{Feature: "f"}, 99, "always"), "g1-file-size") {
		t.Fatal("empty failures must not produce a pre-warning question")
	}
}
