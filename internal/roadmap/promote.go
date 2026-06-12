package roadmap

import (
	"encoding/json"
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

// Promote moves a scored Backlog finding into the target phase and appends the
// analysis + quality artifacts (raw-preserving). Scores must already be
// validated by ParseScores. Roadmap.json is written only after the finding is
// located and the target phase is confirmed to exist.
func Promote(path string, req PromoteRequest) (*BacklogFinding, error) {
	doc, err := readRawRoadmap(path)
	if err != nil {
		return nil, err
	}
	raw, backlogIdx, err := doc.findInBacklog(req.Slug)
	if err != nil {
		return nil, err
	}
	var finding BacklogFinding
	if err := json.Unmarshal(raw, &finding); err != nil {
		return nil, err
	}
	// Pre-flight: appendToPhase mutates only in-memory (sets dirty), and the
	// artifact files are validated, BEFORE the first byte hits disk. Any failure
	// here leaves all five files byte-identical (no half-promoted state).
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
