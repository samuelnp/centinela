package main

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// resolveEmitModel resolves the driver model id stamped onto a telemetry event.
// Telemetry stays a config-only leaf: the VALUE is resolved here at the cmd/
// emit site and passed in. Precedence: a loaded workflow's pinned DriverModel
// (the model that actually drove the work), else config.DriverModelFrom (env /
// [orchestration] driver_model), else "".
func resolveEmitModel(wf *workflow.Workflow, cfg *config.Config) string {
	if wf != nil && wf.DriverModel != "" {
		return wf.DriverModel
	}
	return config.DriverModelFrom("", cfg)
}

// resolveEmitModelFrom picks the first active workflow's model for emit sites
// that hold a workflow set rather than a single loaded workflow (the prewrite
// hook). With no active workflow it falls back to the env/config model.
func resolveEmitModelFrom(wfs []*workflow.Workflow, cfg *config.Config) string {
	for _, wf := range wfs {
		if wf != nil && wf.DriverModel != "" {
			return wf.DriverModel
		}
	}
	return config.DriverModelFrom("", cfg)
}
