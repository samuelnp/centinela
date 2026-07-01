package roadmap

import "testing"

// canonicalView is the 4-feature Q1 roadmap used across BuildView tests:
// auth-service (done), billing-api (in-progress), checkout-ui (planned→ready),
// reporting (planned→blocked by the in-progress billing-api).
func canonicalView() *Roadmap {
	return &Roadmap{Phases: []Phase{{Name: "Q1", Features: []Feature{
		{Name: "auth-service"},
		{Name: "billing-api"},
		{Name: "checkout-ui", DependsOn: []string{"auth-service"}},
		{Name: "reporting", DependsOn: []string{"billing-api"}},
	}}}}
}

func viewFeature(t *testing.T, v RoadmapView, name string) FeatureView {
	t.Helper()
	for _, p := range v.Phases {
		for _, f := range p.Features {
			if f.Name == name {
				return f
			}
		}
	}
	t.Fatalf("feature %q not found in view", name)
	return FeatureView{}
}

// BuildView maps status/readiness/blockedBy and preserves declared order.
func TestBuildView_StatusReadinessMapping(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "auth-service")
	seedStep(t, "billing-api", "code")
	v := BuildView(canonicalView())
	if len(v.Phases) != 1 || v.Phases[0].Name != "Q1" || len(v.Phases[0].Features) != 4 {
		t.Fatalf("want one Q1 phase with 4 features, got %+v", v.Phases)
	}
	order := []string{"auth-service", "billing-api", "checkout-ui", "reporting"}
	for i, f := range v.Phases[0].Features {
		if f.Name != order[i] {
			t.Fatalf("order[%d]=%q want %q", i, f.Name, order[i])
		}
	}
	if a := viewFeature(t, v, "auth-service"); a.Status != "done" || a.Readiness != "" || a.BlockedBy != nil {
		t.Fatalf("done row must carry status only: %+v", a)
	}
	if b := viewFeature(t, v, "billing-api"); b.Status != "in-progress" || b.Readiness != "" || b.BlockedBy != nil {
		t.Fatalf("in-progress row must omit readiness: %+v", b)
	}
	if c := viewFeature(t, v, "checkout-ui"); c.Status != "planned" || c.Readiness != "ready" || c.BlockedBy != nil {
		t.Fatalf("ready row must omit blockedBy: %+v", c)
	}
	r := viewFeature(t, v, "reporting")
	if r.Status != "planned" || r.Readiness != "blocked" || len(r.BlockedBy) != 1 || r.BlockedBy[0] != "billing-api" {
		t.Fatalf("blocked row must name its unmet dep: %+v", r)
	}
}

// BuildView tallies schedulable counts and always emits a non-nil DependsOn.
func TestBuildView_CountsAndDependsOn(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "auth-service")
	seedStep(t, "billing-api", "code")
	v := BuildView(canonicalView())
	if v.Counts != (StatusCounts{Planned: 2, InProgress: 1, Done: 1}) {
		t.Fatalf("counts = %+v", v.Counts)
	}
	if a := viewFeature(t, v, "auth-service"); a.DependsOn == nil || len(a.DependsOn) != 0 {
		t.Fatalf("no-dep row must serialize dependsOn as empty non-nil slice: %+v", a.DependsOn)
	}
	if c := viewFeature(t, v, "checkout-ui"); len(c.DependsOn) != 1 || c.DependsOn[0] != "auth-service" {
		t.Fatalf("dependsOn must preserve declared deps: %+v", c.DependsOn)
	}
}
