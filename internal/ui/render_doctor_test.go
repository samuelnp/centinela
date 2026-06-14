package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/doctor"
)

func TestRenderDiagnosisGlyphs(t *testing.T) {
	cases := []struct {
		status doctor.Status
		glyph  string
	}{
		{doctor.OK, "✓"},
		{doctor.Warn, "⚠"},
		{doctor.Error, "✗"},
	}
	for _, c := range cases {
		out := RenderDiagnosis(doctor.Diagnosis{Name: "x", Status: c.status, Message: "m"})
		if !strings.Contains(out, c.glyph) || !strings.Contains(out, "x") {
			t.Errorf("status %v: missing glyph %q or name in %q", c.status, c.glyph, out)
		}
	}
}

func TestRenderDiagnosisDetailsAndCommand(t *testing.T) {
	d := doctor.Diagnosis{
		Name:    "worktrees",
		Status:  doctor.Error,
		Message: "abandoned",
		Details: []string{"gone: git worktree remove .worktrees/gone"},
		Repair:  &doctor.Repair{Command: "git worktree remove .worktrees/gone"},
	}
	out := RenderDiagnosis(d)
	if !strings.Contains(out, "·") {
		t.Fatal("details must render with a bullet")
	}
	if !strings.Contains(out, "→ run: git worktree remove") {
		t.Fatalf("report-only command must render: %q", out)
	}
}

func TestRenderDiagnosisNoCommandWhenSafeRepair(t *testing.T) {
	d := doctor.Diagnosis{
		Name: "hooks", Status: doctor.Error, Message: "m",
		Repair: &doctor.Repair{Safe: true, Apply: func() error { return nil }},
	}
	if strings.Contains(RenderDiagnosis(d), "→ run:") {
		t.Fatal("safe repair (no Command) must not render a run line")
	}
}

func TestRenderDoctorSummary(t *testing.T) {
	out := RenderDoctorSummary([]doctor.Diagnosis{
		{Status: doctor.OK}, {Status: doctor.OK}, {Status: doctor.Warn}, {Status: doctor.Error},
	})
	if !strings.Contains(out, "2 ok, 1 warn, 1 error") {
		t.Fatalf("summary: %q", out)
	}
}
