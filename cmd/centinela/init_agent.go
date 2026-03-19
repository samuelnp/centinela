package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/setup"
	"github.com/samuelnp/centinela/internal/ui"
)

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

func isValidAgent(agent string) bool {
	return agent == "claude" || agent == "opencode" || agent == "both"
}

func usesClaude(agent string) bool {
	return agent == "claude" || agent == "both"
}

func usesOpenCode(agent string) bool {
	return agent == "opencode" || agent == "both"
}
