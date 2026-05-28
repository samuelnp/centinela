package evidence

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestSetTopLevelScalars(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := SetField(s, "status", "done"); err != nil {
		t.Fatal(err)
	}
	if s.Status != "done" {
		t.Fatalf("status not set: %q", s.Status)
	}
	if err := SetField(s, "handoffTo", "feature-specialist"); err != nil {
		t.Fatal(err)
	}
	if err := SetField(s, "mobileFirst", "true"); err != nil {
		t.Fatal(err)
	}
	if s.MobileFirst == nil || !*s.MobileFirst {
		t.Fatalf("mobileFirst not set: %+v", s.MobileFirst)
	}
}

func TestSetListFieldsRejected(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	err := SetField(s, "outputs", "docs/plans/alpha.md")
	if err == nil || !strings.Contains(err.Error(), "list") {
		t.Fatalf("expected list rejection, got %v", err)
	}
}

func TestSetExtraStoresFreeForm(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := SetField(s, "extra.note", "reviewed by sam"); err != nil {
		t.Fatal(err)
	}
	out, _ := s.MarshalJSON()
	if !strings.Contains(string(out), `"note": "reviewed by sam"`) {
		t.Fatalf("extra.note missing: %s", out)
	}
}

func TestSetMetaPaths(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := SetField(s, "_meta.cli_version", "v2"); err != nil {
		t.Fatal(err)
	}
	if s.Meta.CLIVersion != "v2" {
		t.Fatalf("cli_version not set: %q", s.Meta.CLIVersion)
	}
	if err := SetField(s, "_meta.bogus", "x"); err == nil {
		t.Fatal("expected error on unknown meta key")
	}
}

func TestSetUnknownField(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	err := SetField(s, "bogus", "x")
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("expected unknown-field error, got %v", err)
	}
}

func TestParseBoolBranches(t *testing.T) {
	for _, in := range []string{"true", "1", "yes", "TRUE"} {
		if b, err := parseBool(in); err != nil || !b {
			t.Fatalf("parseBool(%q) -> %v,%v", in, b, err)
		}
	}
	for _, in := range []string{"false", "0", "no"} {
		if b, err := parseBool(in); err != nil || b {
			t.Fatalf("parseBool(%q) -> %v,%v", in, b, err)
		}
	}
	if _, err := parseBool("maybe"); err == nil {
		t.Fatal("expected error")
	}
}
