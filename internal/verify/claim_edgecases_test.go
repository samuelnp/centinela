package verify

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
)

func TestCheckEdgeCases(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "tests/unit/x_test.go",
		"package p\nfunc TestMissingEvidenceFiles(t *testing.T){}\nfunc TestMisconfiguredTestCommand(t *testing.T){}\n")

	cases := []struct {
		name      string
		edgeCases []string
		want      Status
		nameIn    string
	}{
		{"matched-pass", []string{"missing evidence files", "misconfigured test command"}, StatusPass, ""},
		{"unmatched-warn", []string{"galaxy supernova explodes"}, StatusWarn, "supernova"},
		{"empty-skip", nil, StatusSkip, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := &evidence.RoleEvidence{EdgeCases: tc.edgeCases}
			got := checkEdgeCases(root, "qa", ev)
			if got.Status != tc.want {
				t.Fatalf("status = %q want %q (detail %q)", got.Status, tc.want, got.Detail)
			}
			if tc.nameIn != "" && !strings.Contains(got.Detail, tc.nameIn) {
				t.Fatalf("warn detail should name the unmatched entry, got %q", got.Detail)
			}
		})
	}
}

func TestEdgeCaseMatches(t *testing.T) {
	hay := "testmissingevidencefiles testtimeoutwhilesuiteruns "
	if !edgeCaseMatches("timeout while suite runs", hay) {
		t.Error("expected a significant-word match")
	}
	// Short words (< 4 chars) are ignored, so a phrase of only short words misses.
	if edgeCaseMatches("a is on", hay) {
		t.Error("short words must not match")
	}
}
