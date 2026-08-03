package workflow

import (
	"strings"
	"testing"
)

// The degraded path must not turn every unparseable JSON in .workflow/ into a
// phantom workflow. roadmap.json and friends carry none of the state markers,
// so a future roadmap schema stays a parse error rather than becoming a
// workflow named "roadmap" that doctor would then offer to `rm`.
func TestFutureVersionInNonStateJSONIsNotAWorkflow(t *testing.T) {
	stateRepo(t)
	writeRawState(t, "roadmap", `{"schemaVersion":"2.0","phases":[{"name":"P0"}]}`)
	if _, err := Load("roadmap"); err == nil {
		t.Fatal("a future-versioned non-state document must not load as a workflow")
	}
}

// A per-role evidence JSON keeps its own feature name, so even when it comes
// from the future it is still filtered out of the active set by the
// feature-equals-file-name rule.
func TestFutureVersionEvidenceJSONKeepsItsFeature(t *testing.T) {
	dir := stateRepo(t)
	writeRawState(t, "delta-qa-senior",
		`{"schemaVersion":"2.0","feature":"delta","role":"qa-senior","steps":[]}`)
	wf, err := Load("delta-qa-senior")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if wf.Feature != "delta" {
		t.Fatalf("Feature = %q, want the salvaged \"delta\"", wf.Feature)
	}
	if got := ActiveWorkflows(dir + "/" + WorkflowDir); len(got) != 0 {
		t.Fatalf("an evidence JSON must never be an active workflow, got %d", len(got))
	}
}

// An unreadable version fails CLOSED in Save: "I cannot parse this file's
// version" must never be read as "version 1", or the permissive Load above
// would hand a clobbering Save the very files it rescued.
func TestUnreadableVersionIsRefusedNotAssumedLegacy(t *testing.T) {
	stateRepo(t)
	writeRawState(t, "delta", `{"schemaVersion":"2.0","feature":"delta","currentStep":"code","steps":{}}`)
	err := Save(New("delta")) // never loaded: the CAS guard cannot help here
	if err == nil {
		t.Fatal("a file whose version cannot be read must not be overwritten")
	}
	for _, want := range []string{FilePath("delta"), "cannot read", "centinela update"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("refusal must name %q, got: %v", want, err)
		}
	}
}

// Corruption is not a future version: a truncated file still reports its parse
// error (deferral doctor-detect-corrupt-state-file owns detecting it), and a
// never-loaded Save may still recreate it.
func TestCorruptFileIsStillAParseError(t *testing.T) {
	stateRepo(t)
	writeRawState(t, "broken", `{"feature":"broken","currentStep":"cod`)
	_, err := Load("broken")
	if err == nil || !strings.Contains(err.Error(), "invalid workflow file") {
		t.Fatalf("a corrupt file must report a parse error, got %v", err)
	}
}

// The version probe's own table: what each on-disk token means, independent of
// whether the rest of the document can be modelled.
func TestStateVersionReadsEveryTokenShape(t *testing.T) {
	cases := []struct {
		raw      string
		want     int
		readable bool
	}{
		{`{"feature":"f"}`, legacyVersion, true},
		{`{"schemaVersion":null}`, legacyVersion, true},
		{`{"schemaVersion":0}`, legacyVersion, true},
		{`{"schemaVersion":1}`, 1, true},
		{`{"schemaVersion":99}`, 99, true},
		{`{"schemaVersion":-1}`, -1, true},
		{`{"schemaVersion":"2.0"}`, 0, false},
		{`{"schemaVersion":1.5}`, 0, false},
		{`{"schemaVersion":true}`, 0, false},
		{`{"schemaVersion":99999999999999999999}`, 0, false},
		{`[1,2,3]`, legacyVersion, true},
		{`{trunc`, legacyVersion, true},
	}
	for _, tc := range cases {
		v, ok := stateVersion([]byte(tc.raw))
		if v != tc.want || ok != tc.readable {
			t.Errorf("stateVersion(%s) = %d,%v; want %d,%v", tc.raw, v, ok, tc.want, tc.readable)
		}
	}
}
