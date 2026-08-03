package config

import "testing"

func cfgProject(global, driver string) *Config {
	c := &Config{}
	c.Workflow.EnforcementProfile = global
	c.Workflow.RawEnforcementProfile = global
	c.Orchestration.DriverModel = driver
	return c
}

// TestProjectDefaultProfile walks every tier in both directions. The tail is
// guided; a declared-but-unmapped driver is the one case that falls to strict.
func TestProjectDefaultProfile(t *testing.T) {
	cases := []struct {
		name string
		cfg  *Config
		want string
	}{
		{"nil config", nil, ProfileGuided},
		{"zero config", &Config{}, ProfileGuided},
		{"explicit global strict wins", cfgProject(ProfileStrict, ""), ProfileStrict},
		{"explicit global outranks driver", cfgProject(ProfileStrict, "sonnet"), ProfileStrict},
		{"capable driver → guided", cfgProject("", "sonnet"), ProfileGuided},
		{"limited driver → strict", cfgProject("", "haiku"), ProfileStrict},
		{"frontier driver → outcome", cfgProject("", "opus"), ProfileOutcome},
		{"declared but unmapped driver → strict", cfgProject("", "who/knows"), ProfileStrict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CENTINELA_MODEL", "") // keep the ambient env out of the tier
			if got := ProjectDefaultProfile(tc.cfg); got != tc.want {
				t.Fatalf("ProjectDefaultProfile = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestRequireRoadmapGrading: strict blocks on the grading rungs, the two
// relaxed profiles advise. This knob is PROCESS — no gate reads it.
func TestRequireRoadmapGrading(t *testing.T) {
	if !ProfileDefaults(ProfileStrict).RequireRoadmapGrading {
		t.Fatal("strict must keep the full greenfield cascade")
	}
	for _, p := range []string{ProfileGuided, ProfileOutcome} {
		if ProfileDefaults(p).RequireRoadmapGrading {
			t.Fatalf("%s must demote the grading rungs to advisory", p)
		}
	}
	if !ProfileDefaults("nonsense").RequireRoadmapGrading {
		t.Fatal("an unknown profile normalizes to strict and must keep grading")
	}
}
