package workflow

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRewindToRejections(t *testing.T) {
	cases := []struct {
		name, current, target, reason, wantErr string
	}{
		{"empty-reason", "validate", "code", "  ", "reason must not be empty"},
		{"unknown-step", "validate", "deploy", "x", "unrecognised step"},
		{"forward-target", "code", "tests", "x", "not strictly before"},
		{"equal-target", "validate", "validate", "x", "not strictly before"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wf := New("f")
			wf.CurrentStep = tc.current
			before := len(wf.Revisions)
			if _, err := wf.RewindTo(tc.target, tc.reason); err == nil ||
				!strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want %q", err, tc.wantErr)
			}
			if len(wf.Revisions) != before {
				t.Fatal("state must not mutate on rejection")
			}
		})
	}
}

func TestRewindToDoneRejected(t *testing.T) {
	wf := New("f")
	wf.CurrentStep = "done"
	_, err := wf.RewindTo("code", "reopen")
	if err == nil || !strings.Contains(err.Error(), "completed workflow") {
		t.Fatalf("done err = %v", err)
	}
}

func TestRevisionsRoundTrip(t *testing.T) {
	wf := New("f")
	wf.Revisions = []Revision{{From: "validate", To: "code", Reason: "bug"}}
	data, err := json.Marshal(wf)
	if err != nil {
		t.Fatal(err)
	}
	var got Workflow
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Revisions) != 1 || got.Revisions[0].Reason != "bug" {
		t.Fatalf("round-trip = %+v", got.Revisions)
	}
	// Empty Revisions are omitted from JSON (back-compat with old workflows).
	fresh, _ := json.Marshal(New("f"))
	if strings.Contains(string(fresh), "revisions") {
		t.Fatal("empty revisions must be omitted")
	}
}
