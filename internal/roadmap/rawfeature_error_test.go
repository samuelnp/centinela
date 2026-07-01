package roadmap

import "testing"

// badFeatureBody has a malformed (non-object) feature entry that survives the
// top-level parse but fails per-feature decoding, exercising the error branches.
const badFeatureBody = `{"phases":[{"name":"P","features":[123]}]}`

// TestFindFeature_DecodeError surfaces a per-feature decode failure.
func TestFindFeature_DecodeError(t *testing.T) {
	if _, _, _, err := docFrom(t, badFeatureBody).findFeature("x"); err == nil {
		t.Fatal("malformed feature must error in findFeature")
	}
	if _, err := docFrom(t, badFeatureBody).featurePhase("x"); err == nil {
		t.Fatal("malformed feature must error in featurePhase")
	}
}

// TestFeatureDependents_DecodeError surfaces the raw-scan decode failure.
func TestFeatureDependents_DecodeError(t *testing.T) {
	if _, err := docFrom(t, badFeatureBody).featureDependents("x"); err == nil {
		t.Fatal("malformed feature must error in featureDependents")
	}
	if err := docFrom(t, badFeatureBody).requireNoDependents("x"); err == nil {
		t.Fatal("requireNoDependents must propagate the decode error")
	}
}

// TestToRoadmap_DecodeError surfaces the typed-decode failure.
func TestToRoadmap_DecodeError(t *testing.T) {
	if _, err := docFrom(t, badFeatureBody).toRoadmap(); err == nil {
		t.Fatal("malformed feature must error in toRoadmap")
	}
}

// TestRemove_DecodeError propagates the findFeature decode error (nothing written).
func TestRemove_DecodeError(t *testing.T) {
	p := crudWrite(t, badFeatureBody)
	if err := Remove(p, "x"); err == nil {
		t.Fatal("Remove must surface the decode error")
	}
}

// TestRemove_ReadError surfaces the readRawRoadmap failure on a missing file.
func TestRemove_ReadError(t *testing.T) {
	if err := Remove(crudWrite(t, crudBody)+".absent", "x"); err == nil {
		t.Fatal("Remove on a missing roadmap.json must error")
	}
}

// badPhaseBody has a phase whose "features" is not an array: the top-level parse
// succeeds but decodePhase fails, exercising the phase-decode error branches.
const badPhaseBody = `{"phases":[{"name":"P","features":"x"}]}`

// TestPhaseDecodeErrors propagates decodePhase failures through the raw helpers.
func TestPhaseDecodeErrors(t *testing.T) {
	if _, _, _, err := docFrom(t, badPhaseBody).findFeature("x"); err == nil {
		t.Fatal("findFeature must surface a phase-decode error")
	}
	if _, err := docFrom(t, badPhaseBody).featurePhase("x"); err == nil {
		t.Fatal("featurePhase must surface a phase-decode error")
	}
	e, _ := compactBytes(Feature{Name: "y"})
	if err := docFrom(t, badPhaseBody).appendFeatureToPhase("P", e); err == nil {
		t.Fatal("appendFeatureToPhase must surface a phase-decode error")
	}
	if err := Add(crudWrite(t, badPhaseBody), AddRequest{Slug: "y", Phase: "P"}); err == nil {
		t.Fatal("Add must surface the phaseFeatureNames decode error")
	}
}

// TestAppendFeatureToPhase_BadEntry rejects an entry whose name cannot be read.
func TestAppendFeatureToPhase_BadEntry(t *testing.T) {
	if err := docFrom(t, crudBody).appendFeatureToPhase("Phase 1: Foundations", []byte(`123`)); err == nil {
		t.Fatal("a non-object entry must error on featureName")
	}
}
