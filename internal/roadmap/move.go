package roadmap

// MoveRequest carries a resolved `roadmap move` request: relocate a feature to a
// different schedulable phase, optionally anchored before/after a sibling there.
// At most one of BeforeAnchor/AfterAnchor is set (enforced by the command); with
// neither, the feature appends at the end of the target phase.
type MoveRequest struct {
	Slug         string
	ToPhase      string
	BeforeAnchor string
	AfterAnchor  string
}

// Move relocates a feature to another schedulable phase via raw-preserving
// read-modify-write. Backlog and Baseline are refused as source or target. The
// feature's bytes (draft flag included) move verbatim, so name-keyed draft and
// quality entries survive unchanged. Untouched phases round-trip byte-identically
// and a rejected move (unknown phase/anchor, cycle) writes nothing.
func Move(path string, req MoveRequest) error {
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	entry, srcIdx, _, err := doc.findFeature(req.Slug)
	if err != nil {
		return err
	}
	if err := doc.requireSchedulablePhaseIdx(srcIdx); err != nil {
		return err
	}
	targetIdx, err := doc.schedulablePhaseIndex(req.ToPhase)
	if err != nil {
		return err
	}
	if err := doc.removeFeatureAt(srcIdx, req.Slug); err != nil {
		return err
	}
	pos, err := doc.anchorPos(targetIdx, req.BeforeAnchor, req.AfterAnchor)
	if err != nil {
		return err
	}
	if err := doc.insertFeatureAt(targetIdx, pos, entry); err != nil {
		return err
	}
	return finalizeMutation(path, doc)
}
