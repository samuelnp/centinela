package config

import (
	"fmt"
	"strings"
)

// RoadmapDriftConfig controls the roadmap-drift gate. When Enabled, the gate
// regenerates ROADMAP.md in memory from roadmap.json and byte-compares it with
// the on-disk file. Severity ("fail"|"warn") is the adoption knob: warn
// surfaces drift without blocking merge, fail rejects it.
type RoadmapDriftConfig struct {
	Enabled  bool   `toml:"enabled"`
	Severity string `toml:"severity"`
}

// NormalizeRoadmapDrift trims whitespace and defaults an unset severity to warn
// (the safe-adoption default; operators ratchet to fail once clean).
func NormalizeRoadmapDrift(c RoadmapDriftConfig) RoadmapDriftConfig {
	c.Severity = strings.TrimSpace(c.Severity)
	if c.Severity == "" {
		c.Severity = "warn"
	}
	return c
}

// validateRoadmapDrift rejects an unknown severity so a typo cannot silently
// change the gate's strictness. It is a no-op when the gate is disabled.
func validateRoadmapDrift(c RoadmapDriftConfig) error {
	if !c.Enabled {
		return nil
	}
	if c.Severity != "fail" && c.Severity != "warn" {
		return fmt.Errorf("gates.roadmap_drift.severity must be fail or warn, got %q", c.Severity)
	}
	return nil
}
