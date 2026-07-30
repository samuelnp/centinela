package acceptance

import "testing"

// One parser serves cucumber-js and godog alike: godog copies cucumber's
// summary line.
func TestDetect_CucumberSummaries(t *testing.T) {
	cases := []struct {
		name                                       string
		output                                     string
		scenarios, passed, skipped, pending, undef int
	}{
		{"skipped", "3 scenarios (1 skipped, 2 passed)\n", 3, 2, 1, 0, 0},
		{"godog-undefined", "2 scenarios (2 undefined)\n", 2, 0, 0, 0, 2},
		{"pending", "4 scenarios (1 pending, 3 passed)\n", 4, 3, 0, 1, 0},
		{"all-passed", "5 scenarios (5 passed)\n", 5, 5, 0, 0, 0},
		{"zero-bare", "0 scenarios\n", 0, 0, 0, 0, 0},
		{"zero-breakdown", "0 scenarios (0 passed)\n", 0, 0, 0, 0, 0},
		{"singular", "1 scenario (1 passed)\n", 1, 1, 0, 0, 0},
		{"crlf", "3 scenarios (1 skipped, 2 passed)\r\n", 3, 2, 1, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, ok := Detect("Feature: x\n\n" + c.output + "8 steps (8 passed)\n")
			if !ok || s.Shape != ShapeCucumber || !s.SkipData {
				t.Fatalf("expected a cucumber summary with skip data, got %+v ok=%v", s, ok)
			}
			got := []int{s.Scenarios, s.Passed, s.Skipped, s.Pending, s.Undefined}
			want := []int{c.scenarios, c.passed, c.skipped, c.pending, c.undef}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("counts = %v, want %v", got, want)
				}
			}
		})
	}
}

// The aggregate summary is the LAST one: a per-feature run prints several.
func TestDetect_CucumberUsesTheFinalSummary(t *testing.T) {
	out := "2 scenarios (2 passed)\n\n5 scenarios (1 skipped, 4 passed)\n"
	s, ok := Detect(out)
	if !ok || s.Scenarios != 5 || s.Skipped != 1 {
		t.Fatalf("expected the aggregate summary, got %+v ok=%v", s, ok)
	}
}

// E8 false-positive pin: a summary-SHAPED string inside a log line or a failure
// message is not a run summary, and must not produce a skip verdict.
func TestDetect_SummaryShapedTextInOutputIsNotASummary(t *testing.T) {
	cases := map[string]string{
		"mid-line":  "INFO parser produced 2 scenarios (1 skipped) from the feature file\n",
		"prefixed":  "    log.go:12: 2 scenarios (1 skipped)\n",
		"suffixed":  "2 scenarios (1 skipped) <- from the fixture\n",
		"no-anchor": "expected 3 scenarios (1 skipped, 2 passed) but saw none\n",
	}
	for name, out := range cases {
		t.Run(name, func(t *testing.T) {
			if s, ok := Detect(out); ok && s.Shape == ShapeCucumber {
				t.Fatalf("must not match a non-summary line, got %+v", s)
			}
		})
	}
}

// E7: a truncated report is undetermined, never a skip verdict.
func TestDetect_TruncatedReportIsUndetermined(t *testing.T) {
	for _, out := range []string{"", "   \n", "Feature: checkout\n  Scenario: pay\n2 scenar"} {
		if s, ok := Detect(out); ok {
			t.Fatalf("truncated output must be undetermined, got %+v", s)
		}
	}
}
