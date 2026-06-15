package calibration

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestStrictnessRank — strict(2) > guided(1) > outcome(0); unknown ranks strict.
func TestStrictnessRank(t *testing.T) {
	cases := map[string]int{
		config.ProfileStrict: 2, config.ProfileGuided: 1,
		config.ProfileOutcome: 0, "weird": 2, "": 2,
	}
	for p, want := range cases {
		if got := strictnessRank(p); got != want {
			t.Fatalf("rank(%q) = %d, want %d", p, got, want)
		}
	}
}

// TestTighterLooserClamp — tighter/looser step one level and clamp at the ends.
func TestTighterLooserClamp(t *testing.T) {
	if tighter(config.ProfileOutcome) != config.ProfileGuided ||
		tighter(config.ProfileGuided) != config.ProfileStrict || tighter(config.ProfileStrict) != "" {
		t.Fatal("tighter chain/clamp wrong")
	}
	if looser(config.ProfileStrict) != config.ProfileGuided ||
		looser(config.ProfileGuided) != config.ProfileOutcome || looser(config.ProfileOutcome) != "" {
		t.Fatal("looser chain/clamp wrong")
	}
}
