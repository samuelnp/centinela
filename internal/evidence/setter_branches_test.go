package evidence

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestSetEveryTopLevelScalar(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	pairs := []struct{ field, value string }{
		{"feature", "alpha"}, {"step", "plan"}, {"role", "big-thinker"},
		{"status", "done"}, {"generatedAt", "2026-05-12T00:00:00Z"},
		{"handoffTo", "feature-specialist"},
	}
	for _, p := range pairs {
		if err := SetField(s, p.field, p.value); err != nil {
			t.Errorf("set %s: %v", p.field, err)
		}
	}
}

func TestSetMobileFirstFromInvalidBool(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleUXUISpecialist, "v1")
	err := SetField(s, "mobileFirst", "maybe")
	if err == nil || !strings.Contains(err.Error(), "bool") {
		t.Fatalf("expected bool parse error, got %v", err)
	}
}

func TestSetMetaWrittenAt(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := SetField(s, "_meta.written_at", "2026-05-12T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if s.Meta.WrittenAt != "2026-05-12T00:00:00Z" {
		t.Fatalf("written_at not stored: %q", s.Meta.WrittenAt)
	}
}

func TestSetMetaCreatesWhenNil(t *testing.T) {
	r := &RoleEvidence{}
	if err := SetField(r, "_meta.cli_version", "v9"); err != nil {
		t.Fatal(err)
	}
	if r.Meta == nil || r.Meta.CLIVersion != "v9" {
		t.Fatalf("meta not created: %+v", r.Meta)
	}
}

func TestSetExtraCreatesMapWhenNil(t *testing.T) {
	r := &RoleEvidence{}
	if err := SetField(r, "extra.note", "n"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Extra["note"]; !ok {
		t.Fatal("extra map not created")
	}
}

func TestSetFieldNilEvidence(t *testing.T) {
	if err := SetField(nil, "status", "done"); err == nil {
		t.Fatal("expected nil-evidence error")
	}
}

func TestAppendFieldNilEvidence(t *testing.T) {
	if err := AppendField(nil, "outputs", "x"); err == nil {
		t.Fatal("expected nil-evidence error")
	}
}

func TestReadFieldNilEvidence(t *testing.T) {
	if _, err := ReadField(nil, "outputs"); err == nil {
		t.Fatal("expected nil-evidence error")
	}
}
