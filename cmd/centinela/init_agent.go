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
	case "claude":
		return setupClaude()
	default:
		return nil
	}
}

func setupOpenCode() error {
	if changed, err := setup.InjectOpenCodeConfig("opencode.json"); err != nil {
		return fmt.Errorf("failed to update opencode.json: %w", err)
	} else if changed {
		fmt.Println(ui.RenderSuccess("configured opencode.json"))
	} else {
		fmt.Println(ui.StyleMuted.Render("opencode.json already configured"))
	}
	if changed, err := setup.EnsureOpenCodePlugin(); err != nil {
		return fmt.Errorf("failed to write OpenCode plugin: %w", err)
	} else if changed {
		fmt.Println(ui.RenderSuccess("created .opencode/plugins/centinela.js"))
	}
	if changed, err := setup.EnsureAgentsFile(); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	} else if changed {
		fmt.Println(ui.RenderSuccess("created AGENTS.md"))
	}
	return nil
}

// setupAider wires Aider's managed files through the registry-driven plan/apply
// path so the managed-marker seam handles create/update/manual-review.
func setupAider() error {
	plan, err := setup.BuildSyncPlan("aider")
	if err != nil {
		return err
	}
	for _, it := range plan.Items {
		if it.Action == setup.SyncManualReview {
			fmt.Println(ui.StyleYellow.Render("⚠ manual-review " + it.Path + " (" + it.Reason + ")"))
		}
	}
	if err := setup.ApplySync(plan); err != nil {
		return fmt.Errorf("failed to write Aider assets: %w", err)
	}
	for _, it := range plan.Items {
		if it.Action != setup.SyncManualReview {
			fmt.Println(ui.RenderSuccess(string(it.Action) + " " + it.Path))
		}
	}
	return nil
}

func isValidAgent(agent string) bool {
	return setup.IsValidAgent(agent)
}

func invalidAgentError(flag string) error {
	return fmt.Errorf("invalid --agent %q (use: %s)", flag, strings.Join(setup.RegisteredAgents(), "|"))
}
