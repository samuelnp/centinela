package config

import "testing"

// TestIsEnabled_DefaultNil — nil pointer (unset) means enabled.
func TestIsEnabled_DefaultNil(t *testing.T) {
	m := MemoryConfig{Enabled: nil}
	if !m.IsEnabled() {
		t.Fatal("expected enabled=true when Enabled pointer is nil (unset)")
	}
}

// TestIsEnabled_ExplicitTrue — explicit true pointer.
func TestIsEnabled_ExplicitTrue(t *testing.T) {
	v := true
	m := MemoryConfig{Enabled: &v}
	if !m.IsEnabled() {
		t.Fatal("expected enabled=true when Enabled=true")
	}
}

// TestIsEnabled_ExplicitFalse — explicit false pointer (SC-12).
func TestIsEnabled_ExplicitFalse(t *testing.T) {
	v := false
	m := MemoryConfig{Enabled: &v}
	if m.IsEnabled() {
		t.Fatal("expected enabled=false when Enabled=false (SC-12)")
	}
}

// TestNormalizeRecallMaxEntries_Default — zero returns default.
func TestNormalizeRecallMaxEntries_Default(t *testing.T) {
	if got := NormalizeRecallMaxEntries(0); got != DefaultRecallMaxEntries {
		t.Fatalf("expected default %d, got %d", DefaultRecallMaxEntries, got)
	}
}

// TestNormalizeRecallMaxEntries_Positive — positive value passes through.
func TestNormalizeRecallMaxEntries_Positive(t *testing.T) {
	if got := NormalizeRecallMaxEntries(3); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

// TestNormalizeRecallMaxEntries_Negative — negative uses default.
func TestNormalizeRecallMaxEntries_Negative(t *testing.T) {
	if got := NormalizeRecallMaxEntries(-1); got != DefaultRecallMaxEntries {
		t.Fatalf("expected default for negative, got %d", got)
	}
}

// TestNormalizeRecallMaxBytes_Default — zero returns default.
func TestNormalizeRecallMaxBytes_Default(t *testing.T) {
	if got := NormalizeRecallMaxBytes(0); got != DefaultRecallMaxBytes {
		t.Fatalf("expected default %d, got %d", DefaultRecallMaxBytes, got)
	}
}

// TestNormalizeRecallMaxBytes_Positive — positive passes through.
func TestNormalizeRecallMaxBytes_Positive(t *testing.T) {
	if got := NormalizeRecallMaxBytes(1024); got != 1024 {
		t.Fatalf("expected 1024, got %d", got)
	}
}

// TestNormalizeRecallMaxBytes_Negative — negative uses default.
func TestNormalizeRecallMaxBytes_Negative(t *testing.T) {
	if got := NormalizeRecallMaxBytes(-5); got != DefaultRecallMaxBytes {
		t.Fatalf("expected default for negative, got %d", got)
	}
}

// TestApplyMemoryDefaults_SetsDefaults — zero config gets defaults.
func TestApplyMemoryDefaults_SetsDefaults(t *testing.T) {
	cfg := &Config{}
	applyMemoryDefaults(cfg)
	if cfg.Memory.RecallMaxEntries != DefaultRecallMaxEntries {
		t.Fatalf("expected default entries cap, got %d", cfg.Memory.RecallMaxEntries)
	}
	if cfg.Memory.RecallMaxBytes != DefaultRecallMaxBytes {
		t.Fatalf("expected default bytes cap, got %d", cfg.Memory.RecallMaxBytes)
	}
}
