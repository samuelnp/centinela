package config

import "strings"

// DefaultBuildCommand is the build command used when [gates.build] is enabled
// without an explicit command field.
const DefaultBuildCommand = "go build ./cmd/centinela"

// BuildTarget is a single cross-compile target: a {GOOS, GOARCH} pair.
type BuildTarget struct {
	GOOS   string `toml:"goos"`
	GOARCH string `toml:"goarch"`
}

// BuildGateConfig controls the cross-platform build gate (G-Build).
type BuildGateConfig struct {
	Enabled bool          `toml:"enabled"`
	Command string        `toml:"command"`
	Targets []BuildTarget `toml:"targets"`
}

// NormalizeBuildGate applies defaults to a BuildGateConfig. When the gate is
// enabled with no command, it falls back to DefaultBuildCommand. Targets with
// blank GOOS or GOARCH are dropped so a malformed table entry cannot produce a
// silently-empty cross-compile invocation.
func NormalizeBuildGate(b BuildGateConfig) BuildGateConfig {
	if strings.TrimSpace(b.Command) == "" {
		b.Command = DefaultBuildCommand
	}
	cleaned := make([]BuildTarget, 0, len(b.Targets))
	for _, t := range b.Targets {
		t.GOOS = strings.TrimSpace(t.GOOS)
		t.GOARCH = strings.TrimSpace(t.GOARCH)
		if t.GOOS == "" || t.GOARCH == "" {
			continue
		}
		cleaned = append(cleaned, t)
	}
	b.Targets = cleaned
	return b
}
