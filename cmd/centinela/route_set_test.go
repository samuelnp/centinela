package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestRunRouteSet_DowngradeWithReasonIsRecordedAndAudited(t *testing.T) {
	routeRepo(t, "plan", dynamicToml)
	routeSetReason = "config-only change"
	out := captureStdout(t, func() {
		if err := runRouteSet(nil, []string{"f", "senior-engineer", "balanced"}); err != nil {
			t.Fatalf("expected the pre-step downgrade to be accepted: %v", err)
		}
	})
	if !strings.Contains(out, "route set: senior-engineer → balanced (was reasoning)") {
		t.Fatalf("success line missing the tier transition: %s", out)
	}
	route, ok := routeOf(t, "senior-engineer")
	if !ok || route.Tier != "balanced" || route.Reason != "config-only change" || route.DecidedAt == "" {
		t.Fatalf("route not persisted with reason + decidedAt: %#v (%v)", route, ok)
	}
	events, err := os.ReadFile(".workflow/telemetry/events.jsonl")
	if err != nil {
		t.Fatalf("expected a telemetry audit line: %v", err)
	}
	for _, want := range []string{`"type":"route-decision"`, `"tier":"balanced"`, `"prevTier":"reasoning"`} {
		if !strings.Contains(string(events), want) {
			t.Fatalf("telemetry missing %s: %s", want, events)
		}
	}
}

func TestRunRouteSet_UpgradeNeedsNoReasonEvenMidStep(t *testing.T) {
	routeRepo(t, "code", dynamicToml)
	os.WriteFile(orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer), []byte("x"), 0644) //nolint:errcheck
	routeSetReason = "cheaper"
	if err := runRouteSet(nil, []string{"f", "senior-engineer", "balanced"}); err == nil {
		t.Fatal("a downgrade after the step started must be refused")
	}
	routeSetReason = ""
	captureStdout(t, func() {
		if err := runRouteSet(nil, []string{"f", "senior-engineer", "reasoning"}); err != nil {
			t.Fatalf("an upgrade must pass without a reason: %v", err)
		}
	})
	route, ok := routeOf(t, "senior-engineer")
	if !ok || route.Tier != "reasoning" || route.Reason != "" {
		t.Fatalf("upgrade not recorded with an empty reason: %#v (%v)", route, ok)
	}
}
