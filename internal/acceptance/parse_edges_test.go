package acceptance

import "testing"

// A count too large to be an int is not a summary — better undetermined than
// a fabricated number.
func TestParseCucumber_UnrepresentableTotalIsNotASummary(t *testing.T) {
	if s, ok := cucumberSummaryOf("999999999999999999999999 scenarios (1 skipped)"); ok {
		t.Fatalf("an unrepresentable total must not parse, got %+v", s)
	}
}

// Unknown or malformed breakdown entries are ignored, never guessed at.
func TestApplyCucumberCounts_IgnoresUnknownAndMalformedEntries(t *testing.T) {
	s, ok := cucumberSummaryOf("6 scenarios (1 failed, 2 passed, ambiguous, 999999999999999999999999 skipped)")
	if !ok {
		t.Fatal("the summary line itself must still parse")
	}
	if s.Passed != 2 {
		t.Fatalf("passed = %d, want 2", s.Passed)
	}
	if s.Skipped != 0 || s.Pending != 0 || s.Undefined != 0 {
		t.Fatalf("unknown/malformed entries must not become counts, got %+v", s)
	}
}

// Lines that look like JSON but are not decodable events are skipped; a run
// with no decodable event at all is not the -json shape.
func TestParseGoJSON_IgnoresUndecodableLines(t *testing.T) {
	if s, ok := parseGoJSON("{not json\n{\"Package\":\"p\"}\n", ScopeAcceptance); ok {
		t.Fatalf("no decodable action means this is not the -json shape, got %+v", s)
	}
	s, ok := parseGoJSON("{broken\n{\"Package\":\"p\"}\n{\"Action\":\"pass\",\"Test\":\"T\"}\n", ScopeAcceptance)
	if !ok || s.Passed != 1 || s.Scenarios != 1 {
		t.Fatalf("decodable events must still be counted, got %+v ok=%v", s, ok)
	}
}

// A failing test is an executed scenario, not a skipped one; unrelated actions
// (run, output, start) are not scenarios at all.
func TestAddGoResult_FailAndUnrelatedActions(t *testing.T) {
	var s Summary
	addGoResult(&s, "FAIL")
	addGoResult(&s, "RUN")
	addGoResult(&s, "OUTPUT")
	if s.Scenarios != 1 || s.Passed != 0 || s.Skipped != 0 {
		t.Fatalf("only the fail action is a scenario, got %+v", s)
	}
}
