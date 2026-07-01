package roadmap

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PromoteRequest carries a scored promote request. Summary overrides the
// finding's own summary when non-empty (used for the quality entry).
type PromoteRequest struct {
	Slug    string
	Phase   string
	Summary string
	Scores  QualityScores
}

// BacklogFinding is the decoded metadata of a Backlog entry, used to build
// provenance bullets and the promoted quality entry.
type BacklogFinding struct {
	Name       string  `json:"name"`
	Summary    string  `json:"summary"`
	Source     *Source `json:"source"`
	DeferredAt string  `json:"deferredAt"`
}

// LoadBacklogFinding returns the decoded Backlog entry for slug, erroring when
// the slug is not in the Backlog phase. Used by the no-scores evaluator path.
func LoadBacklogFinding(path, slug string) (*BacklogFinding, error) {
	doc, err := readRawRoadmap(path)
	if err != nil {
		return nil, err
	}
	raw, _, err := doc.findInBacklog(slug)
	if err != nil {
		return nil, err
	}
	var f BacklogFinding
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// Promote finalizes a scored feature and appends its analysis + quality
// artifacts, branching on the slug's CURRENT location (never a flag):
//   - a Backlog finding → moves it into req.Phase (today's behavior);
//   - a draft already in a schedulable phase → clears the draft flag in place,
//     no move (see promote_inplace.go);
//   - a non-draft, non-Backlog slug → a clear error.
//
// Scores must already be validated by ParseScores. Roadmap.json is written only
// after the mutation is staged in-memory and artifacts pre-flight, so a rejected
// promote leaves all five files byte-identical.
func Promote(path string, req PromoteRequest) (*BacklogFinding, error) {
	doc, err := readRawRoadmap(path)
	if err != nil {
		return nil, err
	}
	if raw, backlogIdx, ferr := doc.findInBacklog(req.Slug); ferr == nil {
		return promoteFromBacklog(path, doc, raw, backlogIdx, req)
	}
	return promoteDraftInPlace(path, doc, req)
}

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
	if err := preflightArtifacts(); err != nil {
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
