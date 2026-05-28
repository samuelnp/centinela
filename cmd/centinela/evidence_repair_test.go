package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestEvidenceRepairRemovesOrphan(t *testing.T) {
	chdirEvidenceTemp(t)
	orphan := evidence.TempPathFor("alpha", orchestration.RoleBigThinker)
	if err := os.WriteFile(orphan, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		if err := runEvidenceRepair(nil, []string{"alpha"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "removed") {
		t.Fatalf("expected removed message, got %q", out)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Fatal("orphan not removed")
	}
}

func TestEvidenceRepairNoop(t *testing.T) {
	chdirEvidenceTemp(t)
	out := captureStdout(t, func() {
		if err := runEvidenceRepair(nil, []string{"alpha"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "no orphaned") {
		t.Fatalf("expected no-orphan message: %q", out)
	}
}
