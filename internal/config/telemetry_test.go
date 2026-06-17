package config

import "testing"

// IsEnabled defaults to true (opt-out) and honors an explicit setting.
func TestTelemetryConfig_IsEnabled(t *testing.T) {
	if !(TelemetryConfig{}).IsEnabled() {
		t.Fatal("nil Enabled should default to true")
	}
	on := true
	if !(TelemetryConfig{Enabled: &on}).IsEnabled() {
		t.Fatal("explicit true should be enabled")
	}
	off := false
	if (TelemetryConfig{Enabled: &off}).IsEnabled() {
		t.Fatal("explicit false should be disabled")
	}
}
