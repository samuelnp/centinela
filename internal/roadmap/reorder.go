package roadmap

import "fmt"

// ReorderRequest carries a resolved `roadmap reorder` request: reposition a
// feature relative to a sibling anchor. The target phase is the anchor's phase
// (usually the feature's own), so a reorder may also cross phases when the anchor
// lives elsewhere. Exactly one of BeforeAnchor/AfterAnchor is set (enforced by
// the command).
type ReorderRequest struct {
	Slug         string
	BeforeAnchor string
	AfterAnchor  string
}

// Reorder repositions a feature via remove+insert, resolving the anchor after the
// removal so indices stay correct. Source and anchor phase must be schedulable. A
// reorder that resolves to the feature's current position leaves roadmap.json
// byte-identical (no write); a rejected reorder writes nothing.
func Reorder(path string, req ReorderRequest) error {
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	entry, srcIdx, _, err := doc.findFeature(req.Slug)
	if err != nil {
		return err
	}
	if req.BeforeAnchor == req.Slug || req.AfterAnchor == req.Slug {
		return nil // anchoring a feature to itself is a no-op (byte-identical)
	}
	if err := doc.requireSchedulablePhaseIdx(srcIdx); err != nil {
		return err
	}
	anchor := req.BeforeAnchor
	if anchor == "" {
		anchor = req.AfterAnchor
	}
	if anchor == "" {
		return fmt.Errorf("reorder requires --before or --after <anchor>")
	}
	before, err := doc.phaseOrder()
	if err != nil {
		return err
	}
	if err := doc.applyReorder(req, srcIdx, anchor, entry); err != nil {
		return err
	}
	after, err := doc.phaseOrder()
	if err != nil {
		return err
	}
	if sameOrder(before, after) {
		return nil
	}
	return finalizeMutation(path, doc)
}
