package main

import "testing"

func TestRunStatusMissingWorkflow(t *testing.T) {
	if err := runStatus(nil, []string{"nope"}); err == nil {
		t.Fatal("expected missing workflow error")
	}
}
