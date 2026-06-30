package workflow

import (
	"strings"
	"testing"
	"time"
)

// wfAtValidate builds a canonical workflow sitting on the validate step with
// every prior step done (each carrying a non-null CompletedAt timestamp).
func wfAtValidate() *Workflow {
	wf := New("f")
	ts := "2026-06-30T00:00:00Z"
	for _, s := range []string{"plan", "code", "tests"} {
		wf.Steps[s] = StepState{Status: "done", CompletedAt: &ts}
	}
	wf.Steps["validate"] = StepState{Status: "in-progress"}
	wf.CurrentStep = "validate"
	return wf
}

func TestRewindToReopensDownstream(t *testing.T) {
	wf := wfAtValidate()
	reopened, err := wf.RewindTo("code", "bug found")
	if err != nil {
		t.Fatalf("RewindTo: %v", err)
	}
	if got := strings.Join(reopened, ","); got != "tests,validate,docs" {
		t.Fatalf("reopened = %v", reopened)
	}
	if wf.CurrentStep != "code" {
		t.Fatalf("current = %q", wf.CurrentStep)
	}
	if wf.Steps["code"].Status != "in-progress" || wf.Steps["code"].CompletedAt != nil {
		t.Fatalf("code = %+v", wf.Steps["code"])
	}
	for _, s := range []string{"tests", "validate", "docs"} {
		if wf.Steps[s].Status != "pending" || wf.Steps[s].CompletedAt != nil {
			t.Fatalf("%s = %+v", s, wf.Steps[s])
		}
	}
	if wf.Steps["plan"].Status != "done" {
		t.Fatal("plan must stay done")
	}
	if len(wf.Revisions) != 1 {
		t.Fatalf("revisions = %d", len(wf.Revisions))
	}
	r := wf.Revisions[0]
	if r.From != "validate" || r.To != "code" || r.Reason != "bug found" {
		t.Fatalf("revision = %+v", r)
	}
	if r.At.IsZero() || time.Since(r.At) > time.Minute {
		t.Fatalf("revision timestamp unset/old: %v", r.At)
	}
}

func TestReopenedStepsCanonicalAndHotfix(t *testing.T) {
	canon := reopenedSteps([]string{"plan", "code", "tests", "validate", "docs"}, "code")
	if strings.Join(canon, ",") != "tests,validate,docs" {
		t.Fatalf("canon = %v", canon)
	}
	hot := reopenedSteps([]string{"code", "tests", "validate"}, "code")
	if strings.Join(hot, ",") != "tests,validate" {
		t.Fatalf("hotfix = %v", hot)
	}
	if reopenedSteps([]string{"plan", "code"}, "missing") != nil {
		t.Fatal("absent target yields nil")
	}
	if got := reopenedSteps([]string{"plan", "code"}, "code"); len(got) != 0 {
		t.Fatalf("last step reopens nothing, got %v", got)
	}
}

func TestRevisionsSummary(t *testing.T) {
	if RevisionsSummary(nil) != "" {
		t.Fatal("nil → empty")
	}
	wf := New("f")
	if RevisionsSummary(wf) != "" {
		t.Fatal("no revisions → empty")
	}
	wf.Revisions = []Revision{
		{From: "validate", To: "code", Reason: "first"},
		{From: "validate", To: "code", Reason: "second"},
	}
	got := RevisionsSummary(wf)
	if !strings.Contains(got, "2") || !strings.Contains(got, "second") {
		t.Fatalf("summary = %q", got)
	}
}
