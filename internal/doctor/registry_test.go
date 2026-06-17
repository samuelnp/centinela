package doctor

import "testing"

func TestExitError(t *testing.T) {
	if ExitError([]Diagnosis{{Status: OK}, {Status: Warn}}) {
		t.Fatal("OK/Warn must not trigger exit error")
	}
	if !ExitError([]Diagnosis{{Status: OK}, {Status: Error}}) {
		t.Fatal("any Error must trigger exit error")
	}
}

func TestCounts(t *testing.T) {
	ok, warn, err := Counts([]Diagnosis{
		{Status: OK}, {Status: OK}, {Status: Warn}, {Status: Error},
	})
	if ok != 2 || warn != 1 || err != 1 {
		t.Fatalf("counts = %d ok, %d warn, %d err", ok, warn, err)
	}
}

// healthyCtx builds a fixture where every check is OK so Run/Fix exercise the
// full registry deterministically.
func healthyCtx(t *testing.T) Context {
	t.Helper()
	dir := repoFixture(t)
	seedSyncedHooks(t)
	seedRoadmap(t, "Phase 1: Core")
	seedMakefile(t, dir, "0.21.1")
	stubGit(t, okGit(""))
	stubVersion(t, func() (string, error) { return "0.21.1\n", nil })
	cfg := configWithTimeout(240)
	return Context{Root: dir, Config: cfg}
}

func TestRunAllChecksHealthy(t *testing.T) {
	ctx := healthyCtx(t)
	diags := Run(ctx)
	if len(diags) != 7 {
		t.Fatalf("expected 7 checks, got %d", len(diags))
	}
	ok, warn, errc := Counts(diags)
	if ok != 7 || warn != 0 || errc != 0 {
		t.Fatalf("healthy must be all OK, got %d ok %d warn %d err", ok, warn, errc)
	}
	// deterministic order.
	names := []string{"hooks", "roadmap", "worktrees", "workflow-state", "evidence", "config", "version"}
	for i, n := range names {
		if diags[i].Name != n {
			t.Fatalf("order[%d]=%q want %q", i, diags[i].Name, n)
		}
	}
}

func TestRunOrderStableAcrossRuns(t *testing.T) {
	ctx := healthyCtx(t)
	a, b := Run(ctx), Run(ctx)
	for i := range a {
		if a[i].Name != b[i].Name {
			t.Fatalf("non-deterministic order at %d", i)
		}
	}
}
