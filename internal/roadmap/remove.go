package roadmap

// Remove deletes a planned feature from the roadmap via raw-preserving
// read-modify-write. It is refused when the feature does not exist, when its
// derived status is in-progress/done, or when another feature (draft or not)
// still depends on it. All guards run before the single atomic write, so a
// rejected remove leaves roadmap.json byte-identical.
func Remove(path, slug string) error {
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	_, phaseIdx, _, err := doc.findFeature(slug)
	if err != nil {
		return err
	}
	if err := requirePlannedStatus(slug); err != nil {
		return err
	}
	if err := doc.requireNoDependents(slug); err != nil {
		return err
	}
	if err := doc.removeFeatureAt(phaseIdx, slug); err != nil {
		return err
	}
	return writeRawRoadmap(path, doc)
}
