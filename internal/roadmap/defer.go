package roadmap

import (
	"time"
)

// DeferOptions carries a resolved defer request. Source is nil when the finding
// has no provenance (e.g. run outside a worktree with no --source flag).
type DeferOptions struct {
	Slug    string
	Summary string
	Source  *Source
	Now     time.Time // capture time; defaults to time.Now() when zero
}

// Defer validates the request, then appends a Backlog finding to roadmap.json
// via raw-preserving read-modify-write. It creates the Backlog phase (as the
// last phase) when absent. Every validation runs before any write; on any
// failure roadmap.json is left untouched.
func Defer(path string, opts DeferOptions) error {
	if err := validateSlug(opts.Slug); err != nil {
		return err
	}
	if err := validateSummary(opts.Summary); err != nil {
		return err
	}
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	existing, err := doc.phaseFeatureNames()
	if err != nil {
		return err
	}
	if err := validateNoCollision(opts.Slug, existing); err != nil {
		return err
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	entry, err := compactBytes(Feature{
		Name:       opts.Slug,
		Summary:    opts.Summary,
		Source:     opts.Source,
		DeferredAt: now.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	if err := doc.appendBacklog(entry); err != nil {
		return err
	}
	return writeRawRoadmap(path, doc)
}
