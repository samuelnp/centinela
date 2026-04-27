package config

import (
	"reflect"
	"testing"
)

func TestUIPathsUsesDefaults(t *testing.T) {
	got := UIPaths(nil)
	if !reflect.DeepEqual(got, defaultUIPaths) {
		t.Fatalf("expected default ui paths %v, got %v", defaultUIPaths, got)
	}
}

func TestUIPathsNormalizesConfiguredValues(t *testing.T) {
	cfg := &Config{Orchestration: OrchestrationConfig{UIPaths: []string{" ./src/ui ", "web//app", ""}}}
	got := UIPaths(cfg)
	want := []string{"src/ui", "web/app"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected normalized ui paths %v, got %v", want, got)
	}
}
