package config

import "testing"

func TestProfileDefaults_PerProfileKnobs(t *testing.T) {
	cases := []struct {
		profile string
		want    ProfileKnobs
	}{
		{ProfileStrict, ProfileKnobs{
			StepGating: true, ConfirmationMode: ConfirmEveryStep,
			RequireSubagentEvidence: true, PlanAdvisorMode: PlanAdvisorAlways,
		}},
		{ProfileGuided, ProfileKnobs{
			StepGating: true, ConfirmationMode: ConfirmAfterPlan,
			RequireSubagentEvidence: false, PlanAdvisorMode: PlanAdvisorMissingInfo,
		}},
		{ProfileOutcome, ProfileKnobs{
			StepGating: false, ConfirmationMode: ConfirmAuto,
			RequireSubagentEvidence: false, PlanAdvisorMode: PlanAdvisorOff,
		}},
	}
	for _, c := range cases {
		if got := ProfileDefaults(c.profile); got != c.want {
			t.Fatalf("ProfileDefaults(%q) = %+v, want %+v", c.profile, got, c.want)
		}
	}
}

func TestProfileDefaults_UnknownMapsToStrict(t *testing.T) {
	if got := ProfileDefaults("nonsense"); got != ProfileDefaults(ProfileStrict) {
		t.Fatalf("unknown profile must default to strict knobs, got %+v", got)
	}
}

// Only strict requires subagent evidence — the welded-on verification axis is
// unaffected, but process evidence is the knob that distinguishes strict.
func TestProfileDefaults_OnlyStrictRequiresSubagentEvidence(t *testing.T) {
	if !ProfileDefaults(ProfileStrict).RequireSubagentEvidence {
		t.Fatal("strict must require subagent evidence")
	}
	if ProfileDefaults(ProfileGuided).RequireSubagentEvidence ||
		ProfileDefaults(ProfileOutcome).RequireSubagentEvidence {
		t.Fatal("guided/outcome must NOT require subagent evidence")
	}
}
