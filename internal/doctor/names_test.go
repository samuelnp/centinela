package doctor

import "testing"

func TestCheckNames(t *testing.T) {
	want := []string{"hooks", "roadmap", "worktrees", "workflow-state", "evidence", "config", "version"}
	all := checks()
	if len(all) != len(want) {
		t.Fatalf("registry size %d want %d", len(all), len(want))
	}
	for i, c := range all {
		if c.Name() != want[i] {
			t.Errorf("checks[%d].Name()=%q want %q", i, c.Name(), want[i])
		}
	}
}
