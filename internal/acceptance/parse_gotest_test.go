package acceptance

import "testing"

const jsonSkip = `{"Action":"run","Package":"p","Test":"TestA"}
{"Action":"skip","Package":"p","Test":"TestA"}
{"Action":"pass","Package":"p","Test":"TestB"}
{"Action":"pass","Package":"p"}
`

// A test-level skip action is a skipped scenario.
func TestDetect_GoJSONTestLevelSkip(t *testing.T) {
	s, ok := Detect(jsonSkip, ScopeAcceptance)
	if !ok || s.Shape != ShapeGoJSON || !s.SkipData {
		t.Fatalf("expected go test -json with skip data, got %+v ok=%v", s, ok)
	}
	if s.Skipped != 1 || s.Passed != 1 || s.Scenarios != 2 {
		t.Fatalf("counts = %+v", s)
	}
}

// The false-positive pin: a PACKAGE-level skip is "[no test files]", not a
// skipped scenario. Package events carry no Test field.
func TestDetect_GoJSONPackageLevelSkipIsNotAScenarioSkip(t *testing.T) {
	out := `{"Action":"skip","Package":"p","Output":"?   \tp\t[no test files]\n"}
{"Action":"pass","Package":"q","Test":"TestOK"}
{"Action":"pass","Package":"q"}
`
	s, ok := Detect(out, ScopeAcceptance)
	if !ok || s.Skipped != 0 {
		t.Fatalf("package-level skip must not count as a scenario skip, got %+v ok=%v", s, ok)
	}
	if s.Unexecuted() != 0 || s.Scenarios != 1 {
		t.Fatalf("expected one executed scenario and no skips, got %+v", s)
	}
}

// E9: parallel packages interleave their -json lines; counts stay correct
// because the format is one self-contained JSON object per line.
func TestDetect_GoJSONInterleavedPackages(t *testing.T) {
	out := `{"Action":"run","Package":"a","Test":"T1"}
{"Action":"run","Package":"b","Test":"T2"}
{"Action":"skip","Package":"b","Test":"T2"}
{"Action":"pass","Package":"a","Test":"T1"}
not json at all
`
	s, ok := Detect(out, ScopeAcceptance)
	if !ok || s.Skipped != 1 || s.Passed != 1 || s.Scenarios != 2 {
		t.Fatalf("interleaved counts wrong: %+v ok=%v", s, ok)
	}
}

// go test -v reports skips as --- SKIP lines, including indented subtests.
func TestDetect_GoVerboseSkip(t *testing.T) {
	out := "=== RUN   TestA\n--- SKIP: TestA (0.00s)\n=== RUN   TestB\n" +
		"    --- SKIP: TestB/sub (0.00s)\n--- PASS: TestB (0.00s)\nPASS\nok  \tp\t0.1s\n"
	s, ok := Detect(out, ScopeAcceptance)
	if !ok || s.Shape != ShapeGoVerbose || !s.SkipData {
		t.Fatalf("expected go test -v with skip data, got %+v ok=%v", s, ok)
	}
	if s.Skipped != 2 || s.Passed != 1 || s.Scenarios != 3 {
		t.Fatalf("counts = %+v", s)
	}
}

// R4: plain non-verbose `go test` is RECOGNIZED but carries no skip data. That
// is a quiet detail, not a warning — the shape can never carry the data, so a
// per-run ⚠ would be permanent and would train operators to ignore ⚠.
func TestDetect_GoNonVerboseIsRecognizedWithoutSkipData(t *testing.T) {
	out := "ok  \tgithub.com/x/a\t0.30s\tcoverage: 97.1% of statements\n" +
		"?   \tgithub.com/x/b\t[no test files]\n" +
		"ok  \tgithub.com/x/c\t(cached)\n"
	s, ok := Detect(out, ScopeAcceptance)
	if !ok {
		t.Fatal("plain go test output must be recognized, not reported unparseable")
	}
	if s.Shape != ShapeGoNonVerbose || s.SkipData {
		t.Fatalf("expected the non-verbose shape without skip data, got %+v", s)
	}
}

// Output from an unsupported runner stays undetermined.
func TestDetect_UnknownRunnerIsUndetermined(t *testing.T) {
	if s, ok := Detect("Ran 12 examples, 0 failures\nFinished in 1.2 seconds\n", ScopeAcceptance); ok {
		t.Fatalf("an unsupported runner must be undetermined, got %+v", s)
	}
}
