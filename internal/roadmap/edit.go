package roadmap

import "encoding/json"

// EditRequest carries a resolved `roadmap edit` request. Only fields the caller
// provides are applied: Description/Archetype apply when non-empty, DependsOn
// applies only when SetDeps is true (the cobra Changed("depends-on") sentinel, so
// an explicit empty --depends-on clears deps while an omitted flag preserves
// them). A non-empty NewName that differs from Slug triggers a rename.
type EditRequest struct {
	Slug        string
	NewName     string
	Description string
	Archetype   string
	DependsOn   []string
	SetDeps     bool
}

// Edit mutates an existing feature in place via raw-preserving read-modify-write.
// Only provided fields change; a rename validates the new slug, refuses a
// collision, and rewrites every dependent's dependsOn. Dependency integrity is
// re-checked before the single atomic write, so a rejected edit (bad slug,
// collision, cycle, unknown dep) leaves roadmap.json byte-identical.
func Edit(path string, req EditRequest) error {
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	raw, phaseIdx, featIdx, err := doc.findFeature(req.Slug)
	if err != nil {
		return err
	}
	if editIsNoop(req) {
		return nil // nothing effectively changes → leave roadmap.json byte-identical
	}
	var feat Feature
	if err := json.Unmarshal(raw, &feat); err != nil {
		return err
	}
	applyEditFields(&feat, req)
	if err := doc.applyRename(&feat, req); err != nil {
		return err
	}
	entry, err := compactBytes(feat)
	if err != nil {
		return err
	}
	if err := doc.replaceFeatureAt(phaseIdx, featIdx, entry); err != nil {
		return err
	}
	return finalizeMutation(path, doc)
}

// editIsNoop reports whether an edit request changes nothing: no rename (empty or
// same-name), no description/archetype, and no --depends-on. Such an edit is a
// byte-identical no-op — the file is not rewritten at all (mirrors reorder's
// order-preserving no-op guard).
func editIsNoop(req EditRequest) bool {
	noRename := req.NewName == "" || req.NewName == req.Slug
	return noRename && req.Description == "" && req.Archetype == "" && !req.SetDeps
}
