package roadmap

import (
	"encoding/json"
	"fmt"
	"strings"
)

// promoteFromBacklog implements the unchanged Backlog move-and-score path.
func promoteFromBacklog(path string, doc *rawDoc, raw json.RawMessage, backlogIdx int, req PromoteRequest) (*BacklogFinding, error) {
	var finding BacklogFinding
	if err := json.Unmarshal(raw, &finding); err != nil {
		return nil, err
	}
	if req.Phase == "" {
		return nil, fmt.Errorf("--phase is required to promote a Backlog finding")
	}
	if err := doc.appendToPhase(req.Phase, req.Slug); err != nil {
		return nil, err // unknown phase / duplicate in target — nothing written yet
	}
	if err := doc.removeBacklogFeature(backlogIdx, req.Slug); err != nil {
		return nil, err
	}
	if err := seedThenPreflight(req); err != nil {
		return nil, err // missing/corrupt artifact — nothing written yet
	}
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		summary = finding.Summary
	}
	if err := writeRawRoadmap(path, doc); err != nil {
		return nil, err
	}
	if err := appendPromotionArtifacts(req.Slug, summary, req.Scores, &finding); err != nil {
		return &finding, err
	}
	return &finding, nil
}
