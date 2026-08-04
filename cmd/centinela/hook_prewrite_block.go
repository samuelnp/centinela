package main

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

func blockPrewrite(d hookpolicy.PrewriteDecision, cfg *config.Config, wfs []*workflow.Workflow) {
	model := resolveEmitModelFrom(wfs, cfg)
	if d.NeedInit {
		telemetry.RecordBlock(cfg, "", "", string(d.FileType), d.Path, "need-init", model)
		fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(d.FileType), "", "—", d.Path))
		fmt.Fprintln(os.Stderr, ui.StyleMuted.Render("Run: centinela start <feature>"))
		exitPrewrite(2)
		return
	}
	if d.StaleBinary {
		telemetry.RecordBlock(cfg, d.Feature, "", string(d.FileType), d.Path, "stale-binary", model)
		fmt.Fprintln(os.Stderr, ui.RenderBlockedStaleBinary(string(d.FileType), d.Feature, d.Path))
		exitPrewrite(2)
		return
	}
	telemetry.RecordBlock(cfg, d.Feature, d.Step, string(d.FileType), d.Path, "out-of-step", model)
	fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(d.FileType), d.Step, d.Feature, d.Path))
	exitPrewrite(2)
}
