package config

import (
	"strings"
	"testing"
)

func TestDynamicRoutingEnabled_DefaultsToStatic(t *testing.T) {
	if DynamicRoutingEnabled(nil) {
		t.Fatal("nil config must not enable dynamic routing")
	}
	for _, raw := range []string{"", "static", " STATIC "} {
		cfg := &Config{}
		cfg.Orchestration.RoutingMode = raw
		if DynamicRoutingEnabled(cfg) {
			t.Fatalf("routing_mode %q must resolve to static", raw)
		}
	}
	for _, raw := range []string{"dynamic", " Dynamic "} {
		cfg := &Config{}
		cfg.Orchestration.RoutingMode = raw
		if !DynamicRoutingEnabled(cfg) {
			t.Fatalf("routing_mode %q must enable dynamic routing", raw)
		}
	}
}

func TestValidateRoutingMode(t *testing.T) {
	for _, raw := range []string{"", "static", "dynamic"} {
		cfg := &Config{}
		cfg.Orchestration.RoutingMode = raw
		if err := validateRoutingMode(cfg); err != nil {
			t.Fatalf("routing_mode %q must be accepted: %v", raw, err)
		}
	}
	cfg := &Config{}
	cfg.Orchestration.RoutingMode = "turbo"
	err := validateRoutingMode(cfg)
	if err == nil {
		t.Fatal("expected an error for an unknown routing mode")
	}
	if got := err.Error(); !strings.Contains(got, "routing_mode") || !strings.Contains(got, "turbo") {
		t.Fatalf("error must name the key and the bad value, got %q", got)
	}
}
