package planadvisor

import (
	"testing"

	"github.com/samuelnp/centinela/internal/insights"
)

func TestFailureSummaryRendersOrderedEntries(t *testing.T) {
	fs := []insights.Count{
		{Key: "g1-file-size", Count: 8},
		{Key: "import-graph", Count: 3},
		{Key: "<none>", Count: 1},
	}
	got := failureSummary(fs)
	want := "g1-file-size (×8), import-graph (×3), <none> (×1)"
	if got != want {
		t.Fatalf("summary = %q, want %q", got, want)
	}
}

func TestFailureSummaryEmptyIsEmptyString(t *testing.T) {
	if got := failureSummary(nil); got != "" {
		t.Fatalf("empty summary = %q, want \"\"", got)
	}
	if got := failureSummary([]insights.Count{}); got != "" {
		t.Fatalf("empty-slice summary = %q, want \"\"", got)
	}
}
