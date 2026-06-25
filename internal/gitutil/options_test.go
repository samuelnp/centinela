package gitutil

import (
	"reflect"
	"testing"
)

// TestDeliveryOptionsMatrix exercises all four rows of the {origin × worktree}
// matrix.
func TestDeliveryOptionsMatrix(t *testing.T) {
	cases := []struct {
		origin, worktree bool
		want             []Option
	}{
		{true, true, []Option{OptionPR, OptionMerge}},
		{true, false, []Option{OptionPR}},
		{false, true, []Option{OptionMerge}},
		{false, false, nil},
	}
	for _, c := range cases {
		got := DeliveryOptions(c.origin, c.worktree)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("DeliveryOptions(%v,%v) = %v, want %v", c.origin, c.worktree, got, c.want)
		}
	}
}

// TestSupports reports membership correctly.
func TestSupports(t *testing.T) {
	opts := []Option{OptionPR}
	if !Supports(opts, OptionPR) {
		t.Fatal("pr should be supported")
	}
	if Supports(opts, OptionMerge) {
		t.Fatal("merge should not be supported")
	}
	if Supports(nil, OptionPR) {
		t.Fatal("empty options support nothing")
	}
}
