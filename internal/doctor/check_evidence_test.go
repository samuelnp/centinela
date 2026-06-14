package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvidenceCheckNoneOK(t *testing.T) {
	repoFixture(t)
	d := evidenceCheck{}.Run(Context{})
	if d.Status != OK {
		t.Fatalf("no tmp files must be OK, got %v", d.Status)
	}
}

func TestEvidenceCheckOrphanErrorAndRepair(t *testing.T) {
	repoFixture(t)
	writeFile(t, ".workflow/feat-qa-senior.json.tmp", "{}")
	writeFile(t, ".workflow/feat-senior-engineer.json.tmp", "{}")
	d := evidenceCheck{}.Run(Context{})
	if d.Status != Error || d.Repair == nil || !d.Repair.Safe {
		t.Fatalf("orphans must Error with safe repair, got %v", d.Status)
	}
	if len(d.Details) != 2 {
		t.Fatalf("both tmp paths must be listed, got %v", d.Details)
	}
	if err := d.Repair.Apply(); err != nil {
		t.Fatalf("apply: %v", err)
	}
	left, _ := filepath.Glob(".workflow/*.json.tmp")
	if len(left) != 0 {
		t.Fatalf("repair must remove all tmp files, left %v", left)
	}
	// idempotent re-run.
	if err := d.Repair.Apply(); err != nil {
		t.Fatalf("idempotent apply: %v", err)
	}
	post := evidenceCheck{}.Run(Context{})
	if post.Status != OK {
		t.Fatalf("post-repair must be OK, got %v", post.Status)
	}
}

func TestFeaturePrefix(t *testing.T) {
	cases := []struct{ in, want string }{
		{"feat-qa-senior.json.tmp", "feat"},
		{"my-feature-senior-engineer.json.tmp", "my-feature"},
		{"bare.json.tmp", "bare"},
	}
	for _, c := range cases {
		if got := featurePrefix(c.in); got != c.want {
			t.Errorf("featurePrefix(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestOrphanedTmpsSorted(t *testing.T) {
	repoFixture(t)
	writeFile(t, ".workflow/b-qa-senior.json.tmp", "{}")
	writeFile(t, ".workflow/a-qa-senior.json.tmp", "{}")
	got := orphanedTmps()
	if len(got) != 2 || filepath.Base(got[0]) != "a-qa-senior.json.tmp" {
		t.Fatalf("orphanedTmps must be sorted, got %v", got)
	}
	_ = os.Remove(got[0])
}
