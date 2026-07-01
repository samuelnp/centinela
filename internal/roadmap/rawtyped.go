package roadmap

import "encoding/json"

// toRoadmap decodes the (possibly mutated) phases into a typed Roadmap so a
// post-mutation ValidateDependencies pass runs against the in-memory result
// before any bytes hit disk. Dirty phases are read through phaseBytes.
func (d *rawDoc) toRoadmap() (*Roadmap, error) {
	r := &Roadmap{}
	for i := range d.phases {
		var p Phase
		if err := json.Unmarshal(d.phaseBytes(i), &p); err != nil {
			return nil, err
		}
		r.Phases = append(r.Phases, p)
	}
	return r, nil
}
