package config

import "testing"

func TestNormalizePlanAdvisorFailureTopN(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{0, DefaultPlanAdvisorFailureTopN},
		{-3, DefaultPlanAdvisorFailureTopN},
		{1, 1},
		{3, 3},
		{5, MaxPlanAdvisorFailureTopN},
		{6, MaxPlanAdvisorFailureTopN},
		{100, MaxPlanAdvisorFailureTopN},
	}
	for _, c := range cases {
		if got := NormalizePlanAdvisorFailureTopN(c.in); got != c.want {
			t.Fatalf("NormalizePlanAdvisorFailureTopN(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}
