package config

import "testing"

func TestNormalizeBuildGate_DefaultCommandWhenBlank(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace", "   "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeBuildGate(BuildGateConfig{Command: tc.in})
			if got.Command != DefaultBuildCommand {
				t.Fatalf("expected default %q, got %q", DefaultBuildCommand, got.Command)
			}
		})
	}
}

func TestNormalizeBuildGate_PreservesExplicitCommand(t *testing.T) {
	got := NormalizeBuildGate(BuildGateConfig{Command: "make build"})
	if got.Command != "make build" {
		t.Fatalf("expected explicit command preserved, got %q", got.Command)
	}
}

func TestNormalizeBuildGate_DropsBlankTargets(t *testing.T) {
	in := BuildGateConfig{
		Command: "go build",
		Targets: []BuildTarget{
			{GOOS: "linux", GOARCH: "amd64"},
			{GOOS: "", GOARCH: "arm64"},
			{GOOS: "darwin", GOARCH: ""},
			{GOOS: "  ", GOARCH: "  "},
			{GOOS: " windows ", GOARCH: " arm64 "},
		},
	}
	got := NormalizeBuildGate(in)
	if len(got.Targets) != 2 {
		t.Fatalf("expected 2 valid targets, got %d: %+v", len(got.Targets), got.Targets)
	}
	if got.Targets[0] != (BuildTarget{GOOS: "linux", GOARCH: "amd64"}) {
		t.Fatalf("unexpected first target: %+v", got.Targets[0])
	}
	if got.Targets[1] != (BuildTarget{GOOS: "windows", GOARCH: "arm64"}) {
		t.Fatalf("expected trimmed windows/arm64, got %+v", got.Targets[1])
	}
}

func TestNormalizeBuildGate_EmptyTargetsStayEmpty(t *testing.T) {
	got := NormalizeBuildGate(BuildGateConfig{Command: "go build"})
	if len(got.Targets) != 0 {
		t.Fatalf("expected no targets, got %d", len(got.Targets))
	}
}
