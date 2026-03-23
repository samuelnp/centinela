package main

import (
	"os"
	"testing"
)

func TestRunStatusAllNoWorkflows(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck
	if err := runStatusAll(nil, nil); err != nil {
		t.Fatalf("runStatusAll should pass with no workflows: %v", err)
	}
}
