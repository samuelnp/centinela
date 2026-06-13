package main

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/projectstage"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

// resolveArchetypeOrder picks the step order and the resolved archetype name by
// precedence: an explicit --archetype flag wins, else the roadmap Feature
// archetype, else the bootstrap/canonical order from workflowOrderForFeature.
// An empty/unknown flag or roadmap value is validated before use. The bootstrap
// branch is itself a kind of archetype: it only applies when no archetype is
// explicitly chosen.
func resolveArchetypeOrder(feature, flag string) ([]string, string, error) {
	if name := strings.TrimSpace(flag); name != "" {
		return archetypeOrderByName(name)
	}
	r, err := roadmap.Load()
	if err == nil {
		if name := roadmap.FeatureArchetype(r, feature); strings.TrimSpace(name) != "" {
			return archetypeOrderByName(name)
		}
	}
	order, err := workflowOrderForFeature(feature)
	if err != nil {
		return nil, "", err
	}
	return order, workflow.ArchetypeCanonical, nil
}

// archetypeOrderByName validates a named archetype and returns its step order
// and normalized name.
func archetypeOrderByName(name string) ([]string, string, error) {
	if err := workflow.ValidateArchetype(name); err != nil {
		return nil, "", err
	}
	normalized := workflow.NormalizeArchetype(name)
	order, _ := workflow.ArchetypeStepOrder(normalized)
	return order, normalized, nil
}

func workflowOrderForFeature(feature string) ([]string, error) {
	stage, err := projectstage.Load("PROJECT.md")
	if err != nil {
		return nil, err
	}
	if stage == projectstage.Existing {
		return workflow.DefaultStepOrder, nil
	}
	r, err := roadmap.Load()
	if err != nil {
		return nil, roadmapStartError(err)
	}
	if roadmap.IsBacklogFeature(r, feature) {
		return nil, fmt.Errorf(
			"cannot start %q — it is a Backlog finding; promote it first with "+
				"centinela roadmap promote %s --phase <name>", feature, feature)
	}
	if err := roadmap.ValidateAnalysis(r); err != nil {
		return nil, fmt.Errorf("greenfield project requires roadmap senior PM analysis: %w", err)
	}
	if err := roadmap.ValidateQuality(r); err != nil {
		return nil, fmt.Errorf("greenfield project requires roadmap quality evaluation: %w", err)
	}
	if !roadmap.HasBootstrapPhase(r) {
		return nil, fmt.Errorf("greenfield project requires roadmap phase \"Phase 0: Bootstrap\"")
	}
	if roadmap.IsBootstrapFeature(r, feature) {
		return workflow.BootstrapStepOrder, nil
	}
	if !roadmap.BootstrapComplete(r) {
		return nil, fmt.Errorf("bootstrap is incomplete; start a Phase 0 feature first")
	}
	if err := checkDependencyGuard(r, feature); err != nil {
		return nil, err
	}
	return workflow.DefaultStepOrder, nil
}

func checkDependencyGuard(r *roadmap.Roadmap, feature string) error {
	unmet := roadmap.UnmetDependencies(r, feature)
	if len(unmet) == 0 {
		return nil
	}
	return fmt.Errorf(
		"cannot start %q — blocked by unmet dependencies: %s",
		feature, strings.Join(unmet, ", "),
	)
}
