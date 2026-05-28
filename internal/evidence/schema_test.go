package evidence

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestSkeletonHasContractFields(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if s.Feature != "alpha" || s.Role != "big-thinker" || s.Step != "plan" {
		t.Fatalf("bad skeleton: %+v", s)
	}
	if s.Status != "done" || s.HandoffTo != "feature-specialist" {
		t.Fatalf("bad defaults: %+v", s)
	}
	if s.Meta == nil || s.Meta.CLIVersion != "v1" {
		t.Fatal("meta not stamped")
	}
}

func TestSkeletonUXAddsMobileFirst(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleUXUISpecialist, "v1")
	if s.MobileFirst == nil || !*s.MobileFirst {
		t.Fatalf("expected mobileFirst=true, got %+v", s.MobileFirst)
	}
}

func TestMarshalStableKeyOrder(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	out, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	idxMeta := strings.Index(got, `"_meta"`)
	idxFeature := strings.Index(got, `"feature"`)
	idxHandoff := strings.Index(got, `"handoffTo"`)
	if idxMeta < 0 || idxFeature < 0 || idxHandoff < 0 {
		t.Fatalf("missing keys in %s", got)
	}
	if !(idxMeta < idxFeature && idxFeature < idxHandoff) {
		t.Fatalf("unstable key order: %s", got)
	}
}

func TestRoundTripPreservesUnknownExtra(t *testing.T) {
	raw := []byte(`{"feature":"alpha","step":"plan","role":"big-thinker","status":"done","generatedAt":"2026-05-12T00:00:00Z","inputs":["x"],"outputs":["y"],"edgeCases":[],"handoffTo":"feature-specialist","legacy_field":42}`)
	var r RoleEvidence
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if v, ok := r.Extra["legacy_field"]; !ok || string(v) != "42" {
		t.Fatalf("legacy_field not preserved: %+v", r.Extra)
	}
	out, err := r.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"legacy_field": 42`) {
		t.Fatalf("legacy_field lost on round-trip: %s", out)
	}
}

func TestParseRoleRejectsUnknown(t *testing.T) {
	if _, err := ParseRole("not-a-role"); err == nil {
		t.Fatal("expected error for unknown role")
	}
	if _, err := ParseRole("big-thinker"); err != nil {
		t.Fatalf("known role rejected: %v", err)
	}
}
