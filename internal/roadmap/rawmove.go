package roadmap

import (
	"encoding/json"
	"fmt"
	"strings"
)

// findInBacklog returns the raw bytes of the named finding in the Backlog phase
// and the Backlog phase index, or an error when the slug is not a Backlog entry.
func (d *rawDoc) findInBacklog(slug string) (json.RawMessage, int, error) {
	idx, err := d.backlogPhaseIndex()
	if err != nil {
		return nil, -1, err
	}
	if idx < 0 {
		return nil, -1, fmt.Errorf("%q is not a Backlog finding", slug)
	}
	p, err := d.decodePhase(idx)
	if err != nil {
		return nil, -1, err
	}
	for _, f := range p.Features {
		name, err := featureName(f)
		if err != nil {
			return nil, -1, err
		}
		if name == slug {
			return f, idx, nil
		}
	}
	return nil, -1, fmt.Errorf("%q is not a Backlog finding", slug)
}

// removeBacklogFeature drops the named finding from the Backlog phase.
func (d *rawDoc) removeBacklogFeature(idx int, slug string) error {
	p, err := d.decodePhase(idx)
	if err != nil {
		return err
	}
	kept := p.Features[:0:0]
	for _, f := range p.Features {
		name, _ := featureName(f)
		if name != slug {
			kept = append(kept, f)
		}
	}
	p.Features = kept
	return d.setPhase(idx, p)
}

// appendToPhase appends a name-only feature entry to the named non-Backlog
// phase. Returns an error listing known phases when the target is unknown.
func (d *rawDoc) appendToPhase(target, slug string) error {
	for i := range d.phases {
		p, err := d.decodePhase(i)
		if err != nil {
			return err
		}
		if isBacklogPhaseName(p.Name) || p.Name != target {
			continue
		}
		for _, f := range p.Features {
			if name, _ := featureName(f); name == slug {
				return fmt.Errorf("%q already exists in phase %q; refusing to add a duplicate", slug, target)
			}
		}
		entry, err := compactBytes(Feature{Name: slug, DependsOn: []string{}})
		if err != nil {
			return err
		}
		p.Features = append(p.Features, entry)
		return d.setPhase(i, p)
	}
	return fmt.Errorf("unknown phase %q; known phases: %s", target, d.knownPhaseList())
}

// knownPhaseList returns a comma-joined list of non-Backlog phase names.
func (d *rawDoc) knownPhaseList() string {
	var names []string
	for i := range d.phases {
		if p, err := d.decodePhase(i); err == nil && !isBacklogPhaseName(p.Name) {
			names = append(names, fmt.Sprintf("%q", p.Name))
		}
	}
	return strings.Join(names, ", ")
}
