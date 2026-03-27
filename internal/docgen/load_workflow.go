package docgen

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
)

func loadEvidence() []EvidenceLink {
	paths := listFiles(".workflow/*-*.json")
	out := []EvidenceLink{}
	for _, p := range paths {
		if strings.HasSuffix(p, "roadmap.json") || strings.HasSuffix(p, "analysis.json") {
			continue
		}
		var e struct {
			Role, Feature, Step, HandoffTo string
			Outputs                        []string
			EdgeCases                      []string
		}
		json.Unmarshal([]byte(readFile(p)), &e) //nolint:errcheck
		if e.Role == "" || e.Feature == "" {
			continue
		}
		out = append(out, EvidenceLink{Role: e.Role, Feature: e.Feature, Step: e.Step, Handoff: e.HandoffTo, Outputs: e.Outputs, EdgeCase: len(e.EdgeCases)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Feature+out[i].Role < out[j].Feature+out[j].Role })
	return out
}

func loadStates() []FeatureState {
	paths := listFiles(filepath.Join(".workflow", "*.json"))
	states := []FeatureState{}
	for _, p := range paths {
		if strings.Contains(p, "roadmap") || strings.Count(filepath.Base(p), "-") > 0 {
			continue
		}
		var wf struct {
			Feature, CurrentStep string
			Steps                map[string]struct{ Status string }
		}
		json.Unmarshal([]byte(readFile(p)), &wf) //nolint:errcheck
		if wf.Feature == "" {
			continue
		}
		st := wf.Steps[wf.CurrentStep].Status
		states = append(states, FeatureState{Feature: wf.Feature, Step: wf.CurrentStep, Status: st})
	}
	sort.Slice(states, func(i, j int) bool { return states[i].Feature < states[j].Feature })
	return states
}
