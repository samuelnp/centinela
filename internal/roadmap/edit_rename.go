package roadmap

// applyEditFields overwrites only the fields the request marks as provided.
// Description and Archetype apply when non-empty; DependsOn applies only when
// SetDeps is true, so an omitted --depends-on flag leaves the existing deps
// intact while an explicit (even empty) one replaces them.
func applyEditFields(feat *Feature, req EditRequest) {
	if req.Description != "" {
		feat.Description = req.Description
	}
	if req.Archetype != "" {
		feat.Archetype = req.Archetype
	}
	if req.SetDeps {
		deps := req.DependsOn
		if deps == nil {
			deps = []string{}
		}
		feat.DependsOn = deps
	}
}

// applyRename validates and applies a slug rename: it checks the new slug shape,
// refuses a collision with any existing feature, sets the feature's name, and
// rewrites every dependent's dependsOn from the old slug to the new one. A no-op
// rename (empty or unchanged NewName) returns without touching the doc.
func (d *rawDoc) applyRename(feat *Feature, req EditRequest) error {
	if req.NewName == "" || req.NewName == req.Slug {
		return nil
	}
	if err := validateSlug(req.NewName); err != nil {
		return err
	}
	existing, err := d.phaseFeatureNames()
	if err != nil {
		return err
	}
	if err := validateNoCollision(req.NewName, existing); err != nil {
		return err
	}
	feat.Name = req.NewName
	return d.rewriteDependents(req.Slug, req.NewName)
}
