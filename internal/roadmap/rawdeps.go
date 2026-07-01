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
