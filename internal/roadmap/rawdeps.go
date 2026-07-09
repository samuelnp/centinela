package roadmap

import "encoding/json"

// featureDependents returns the names of every feature that lists slug in its
// dependsOn, across all phases in declared order. Drafts are real dependents:
// this raw scan does not special-case them (only the coverage set does). Used
// as the remove guard so a depended-on feature cannot be deleted.
func (d *rawDoc) featureDependents(slug string) ([]string, error) {
	var out []string
	for i := range d.phases {
		p, err := d.decodePhase(i)
		if err != nil {
			return nil, err
		}
		for _, f := range p.Features {
			var obj struct {
				Name      string   `json:"name"`
				DependsOn []string `json:"dependsOn"`
			}
			if err := json.Unmarshal(f, &obj); err != nil {
				return nil, err
			}
			for _, dep := range obj.DependsOn {
				if dep == slug {
					out = append(out, obj.Name)
					break
				}
			}
		}
	}
	return out, nil
}

// rewriteDependents rewrites every feature's dependsOn entry equal to oldName to
// newName, across all phases. Only phases that actually hold a dependent are
// marked dirty (via setPhase), so untouched phases round-trip byte-identically.
func (d *rawDoc) rewriteDependents(oldName, newName string) error {
	for i := range d.phases {
		p, err := d.decodePhase(i)
		if err != nil {
			return err
		}
		changed := false
		for j, f := range p.Features {
			var feat Feature
			if err := json.Unmarshal(f, &feat); err != nil {
				return err
			}
			hit := false
			for k, dep := range feat.DependsOn {
				if dep == oldName {
					feat.DependsOn[k] = newName
					hit = true
				}
			}
			if !hit {
				continue
			}
			entry, err := compactBytes(feat)
			if err != nil {
				return err
			}
			p.Features[j] = entry
			changed = true
		}
		if changed {
			if err := d.setPhase(i, p); err != nil {
				return err
			}
		}
	}
	return nil
}
