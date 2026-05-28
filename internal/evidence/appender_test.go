package evidence

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestAppendDedupsByEquality(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := AppendField(s, "outputs", "docs/plans/alpha.md"); err != nil {
		t.Fatal(err)
	}
	if err := AppendField(s, "outputs", "docs/plans/alpha.md"); err != nil {
		t.Fatal(err)
	}
	if len(s.Outputs) != 1 {
		t.Fatalf("expected dedup, got %v", s.Outputs)
	}
}

func TestAppendRejectsScalarFields(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	err := AppendField(s, "status", "x")
	if err == nil || !strings.Contains(err.Error(), "not appendable") {
		t.Fatalf("expected non-appendable error, got %v", err)
	}
}

func TestAppendAllListFields(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleFeatureSpecial, "v1")
	if err := AppendField(s, "inputs", "docs/features/alpha.md"); err != nil {
		t.Fatal(err)
	}
	if err := AppendField(s, "outputs", "docs/plans/alpha.md"); err != nil {
		t.Fatal(err)
	}
	if err := AppendField(s, "edgeCases", "migration is safe"); err != nil {
		t.Fatal(err)
	}
	if len(s.Inputs)+len(s.Outputs)+len(s.EdgeCases) != 3 {
		t.Fatalf("expected three entries: %+v", s)
	}
}

func TestReadFieldReturnsTypedValue(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	if err := AppendField(s, "outputs", "docs/plans/alpha.md"); err != nil {
		t.Fatal(err)
	}
	v, err := ReadField(s, "outputs")
	if err != nil {
		t.Fatal(err)
	}
	list, ok := v.([]string)
	if !ok || len(list) != 1 || list[0] != "docs/plans/alpha.md" {
		t.Fatalf("unexpected outputs: %v", v)
	}
	if _, err := ReadField(s, "bogus"); err == nil {
		t.Fatal("expected error on unknown field")
	}
	if _, err := ReadField(s, "extra.missing"); err == nil {
		t.Fatal("expected error on missing extra")
	}
	if err := SetField(s, "extra.note", "hi"); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadField(s, "extra.note"); err != nil {
		t.Fatalf("extra.note should resolve: %v", err)
	}
}

func TestAppendUniqueNoMutationOnHit(t *testing.T) {
	in := []string{"a", "b"}
	out := appendUnique(in, "a")
	if &in[0] != &out[0] || len(out) != 2 {
		t.Fatalf("appendUnique should be no-op on duplicate: %v", out)
	}
}
