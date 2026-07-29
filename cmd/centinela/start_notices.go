package main

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
)

func stepArrow(order []string) string {
	return strings.Join(order, " → ")
}

// printLegacyPlanRoleNotice surfaces the retired [orchestration.models] plan
// keys once, at start. `centinela doctor` reports the same wording as a config
// advisory. The hook path deliberately does NOT print it: it runs on every
// prompt and would turn a one-time migration hint into permanent noise (D8).
func printLegacyPlanRoleNotice(cfg *config.Config) {
	if notice := config.LegacyPlanModelRoleNotice(cfg); notice != "" {
		fmt.Println(ui.StyleMuted.Render(notice))
	}
}
