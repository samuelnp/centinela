package main

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/setup"
	"github.com/samuelnp/centinela/internal/ui"
)

// runHarnessSetup dispatches a registry-resolved harness name to its
// presentation helper. Validity and resolution come from the registry; this
// only renders per-harness UI (which cannot live in internal/setup).
func runHarnessSetup(name string) error {
	switch name {
	case "opencode":
		return setupOpenCode()
	case "aider":
		return setupAider()
	case "codex":
		return setupCodex()
	case "claude":
		return setupClaude()
	default:
		return nil
	}
}

// applyManagedSetup runs the registry-driven managed-sync path for one harness:
// it builds the plan, surfaces manual-review files, applies the writes, and
// renders per-item success. The managed-version header it writes is what the
// migration system expects, so a freshly-init'd project reports no pending
// drift. label names the harness in the failure message only.
func applyManagedSetup(agent, label string) error {
	plan, err := setup.BuildSyncPlan(agent)
	if err != nil {
		return err
	}
	for _, it := range plan.Items {
		if it.Action == setup.SyncManualReview {
			fmt.Println(ui.StyleYellow.Render("⚠ manual-review " + it.Path + " (" + it.Reason + ")"))
		}
	}
	if err := setup.ApplySync(plan); err != nil {
		return fmt.Errorf("failed to write %s assets: %w", label, err)
	}
	for _, it := range plan.Items {
		if it.Action != setup.SyncManualReview {
			fmt.Println(ui.RenderSuccess(string(it.Action) + " " + it.Path))
		}
	}
	return nil
}

func setupOpenCode() error { return applyManagedSetup("opencode", "OpenCode") }

func setupAider() error { return applyManagedSetup("aider", "Aider") }

func setupCodex() error { return applyManagedSetup("codex", "Codex") }

func isValidAgent(agent string) bool {
	return setup.IsValidAgent(agent)
}

func invalidAgentError(flag string) error {
	return fmt.Errorf("invalid --agent %q (use: %s)", flag, strings.Join(setup.RegisteredAgents(), "|"))
}
