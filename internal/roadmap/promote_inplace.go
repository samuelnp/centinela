package roadmap

import (
	"encoding/json"
	"fmt"
	"strings"
)

// promoteDraftInPlace finalizes a draft that already lives in a schedulable
// phase: it clears the draft flag WITHOUT moving the feature, then appends the
// same analysis + quality artifacts as a Backlog promote. A non-draft,
// non-Backlog slug is a clear error. Nothing is written until the mutation is
// staged and artifacts pre-flight, so a rejected finalize is byte-identical.
func promoteDraftInPlace(path string, doc *rawDoc, req PromoteRequest) (*BacklogFinding, error) {
	entry, phaseIdx, featIdx, err := doc.findFeature(req.Slug)
	if err != nil {
		return nil, err
	}
	var f Feature
	if err := json.Unmarshal(entry, &f); err != nil {
		return nil, err
	}
	if !f.Draft {
		return nil, fmt.Errorf(
			"%q is neither a draft nor a Backlog finding; nothing to promote", req.Slug)
	}
	f.Draft = false
	cleared, err := compactBytes(f)
	if err != nil {
		return nil, err
	}
	if err := doc.replaceFeatureAt(phaseIdx, featIdx, cleared); err != nil {
		return nil, err
	}
	if err := preflightArtifacts(); err != nil {
		return nil, err
	}
	summary := draftSummary(req, f)
	if err := writeRawRoadmap(path, doc); err != nil {
		return nil, err
	}
	finding := &BacklogFinding{Name: req.Slug, Summary: summary}
	if err := appendScoreArtifacts(req.Slug, summary, req.Scores, draftFinalizeBullet(req.Slug)); err != nil {
		return finding, err
	}
	return finding, nil
}

// draftSummary resolves the quality-entry summary for a finalized draft: an
// explicit --summary wins, else the feature description, else the slug (the
// quality entry requires a non-empty summary).
func draftSummary(req PromoteRequest, f Feature) string {
	if s := strings.TrimSpace(req.Summary); s != "" {
		return s
	}
	if s := strings.TrimSpace(f.Description); s != "" {
		return s
	}
	return req.Slug
}

// draftFinalizeBullet records the in-place finalize in the artifact .md files.
func draftFinalizeBullet(slug string) string {
	return fmt.Sprintf("- Finalized draft in place: %s", slug)
}
