package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestRunRoadmap(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	if err := runRoadmap(nil, nil); err == nil {
		t.Fatal("expected no roadmap error")
	}
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P1", Features: []roadmap.Feature{{Name: "f"}}}}}
	if err := roadmap.Save(r); err != nil {
		t.Fatalf("save roadmap: %v", err)
	}
	if err := runRoadmap(nil, nil); err != nil {
		t.Fatalf("runRoadmap should pass: %v", err)
	}
}
