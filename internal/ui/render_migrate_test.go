package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/migration"
)

func TestRenderDocsMigrationOutputs(t *testing.T) {
	plan := migration.Plan{Items: []migration.Item{{
		Path: "CLAUDE.md", Action: migration.ActionUpdate,
		FromVersion: "legacy", ToVersion: "1", PreservedKeepBlocks: 1, PreservedCustomSection: 2,
	}}}
	preview := RenderDocsMigrationPlan(plan, false)
	if !strings.Contains(preview, "DOCS PREVIEW") || !strings.Contains(preview, "keep:1 custom:2") {
		t.Fatal("expected preview plan details")
	}
	apply := RenderDocsMigrationPlan(plan, true)
	if !strings.Contains(apply, "DOCS APPLY") {
		t.Fatal("expected apply mode title")
	}
	needed := RenderDocsMigrationNeeded(plan)
	if !strings.Contains(needed, "approval") || !strings.Contains(needed, "centinela migrate docs --apply") {
		t.Fatal("expected migration-needed guidance")
	}
}
