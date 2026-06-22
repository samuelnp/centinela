package reconstruct

import (
	"strings"
	"testing"
)

func TestReconstruct_DeterministicAndTallies(t *testing.T) {
	in := inv("Go", []string{"internal/handler", "internal/service", "cmd/app"})
	a := NewReconstructor().Reconstruct(in)
	b := NewReconstructor().Reconstruct(in)

	if len(a.Targets) != 3 || len(a.Features) != 3 || len(a.Briefs) != 3 {
		t.Fatalf("expected 3 targets/features/briefs, got %d/%d/%d", len(a.Targets), len(a.Features), len(a.Briefs))
	}
	if a.TodoCount != 9 { // 3 features × 3 markers
		t.Fatalf("expected TodoCount 9, got %d", a.TodoCount)
	}
	for i := range a.Features {
		if a.Features[i].Body != b.Features[i].Body || a.Briefs[i].Body != b.Briefs[i].Body {
			t.Fatalf("reconstruct must be byte-identical across runs at %d", i)
		}
		if a.Features[i].Slug != a.Targets[i].Slug {
			t.Fatalf("feature/target slug order must align at %d", i)
		}
	}
}

func TestReconstruct_RoleAwareScenarios(t *testing.T) {
	r := NewReconstructor().Reconstruct(inv("Go", []string{"cmd/app", "internal/handler", "internal/service"}))
	bySlug := map[string]string{}
	for _, f := range r.Features {
		bySlug[f.Slug] = f.Body
	}
	if !strings.Contains(bySlug["cmd-app"], "the command performs its primary behavior") {
		t.Fatalf("command scenario missing:\n%s", bySlug["cmd-app"])
	}
	if !strings.Contains(bySlug["internal-handler"], "the endpoint handles a request") {
		t.Fatalf("endpoint scenario missing:\n%s", bySlug["internal-handler"])
	}
	if !strings.Contains(bySlug["internal-service"], "the module exposes its primary behavior") {
		t.Fatalf("module scenario missing:\n%s", bySlug["internal-service"])
	}
}

func TestReconstruct_EmptyInventoryNoArtifacts(t *testing.T) {
	r := NewReconstructor().Reconstruct(inv("", nil))
	if len(r.Targets) != 0 || len(r.Features) != 0 || r.TodoCount != 0 {
		t.Fatalf("empty inventory must yield no artifacts: %+v", r)
	}
}
