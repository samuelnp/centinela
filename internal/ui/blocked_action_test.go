package ui

import (
	"strings"
	"testing"
)

func TestBlockedAction_BothBranches(t *testing.T) {
	cases := []struct {
		name, step, feature, want string
	}{
		{"empty step", "", "alpha", "centinela start"},
		{"placeholder feature", "plan", "—", "centinela start"},
		{"empty feature", "plan", "", "centinela start"},
		{"happy", "plan", "alpha", "centinela complete alpha"},
	}
	for _, tc := range cases {
		got := blockedAction(tc.step, tc.feature)
		if !strings.Contains(got, tc.want) {
			t.Errorf("%s: blockedAction(%q,%q)=%q want substring %q",
				tc.name, tc.step, tc.feature, got, tc.want)
		}
	}
}
