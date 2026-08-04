package roadmap

import "os"

// seedAnalysisJSON and seedQualityJSON are empty artifact bodies in the exact
// shape writeArtifact emits, so the first append produces no gratuitous diff.
// Both carry "provisional": true — a seeded artifact must never be mistaken for
// an evaluated one, least of all by the strict grading rung (see provisional.go).
const seedAnalysisJSON = "{\n  \"role\": \"senior-product-manager\",\n" +
	"  \"provisional\": true,\n  \"features\": []\n}\n"
const seedQualityJSON = "{\n  \"role\": \"roadmap-quality-evaluator\",\n" +
	"  \"provisional\": true,\n  \"features\": []\n}\n"

const seedAnalysisMD = "# Roadmap Analysis\n\n" +
	"> Seeded by `centinela roadmap promote`. Under the guided/outcome profiles the\n" +
	"> senior-PM analysis is advisory, so this file starts empty and records only the\n" +
	"> features you promote, marked `\"provisional\": true`. Run a real analysis\n" +
	"> pass and clear that mark before pinning `strict`.\n"

const seedQualityMD = "# Roadmap Quality Evaluation\n\n" +
	"> Seeded by `centinela roadmap promote`. Scores are recorded, never gated.\n" +
	"> Under the guided/outcome profiles this evaluation is advisory, so this file\n" +
	"> starts empty and records only the features you promote, marked\n" +
	"> `\"provisional\": true`. Run a real evaluator pass and clear that mark\n" +
	"> before pinning `strict`.\n"

// SeedArtifactsIfAbsent creates the roadmap analysis and quality artifact pairs
// — with their required role headers and an empty features array — for any of
// the four files that does not exist yet. A file that already exists is never
// read, rewritten, or touched.
//
// This is what makes the guided cold start traversable. `start` sends a Backlog
// finding and a draft to `roadmap promote`, and promote records scores by
// APPENDING to those artifacts, so on a tree that has only PROJECT.md and
// .workflow/roadmap.json it would otherwise refuse with "roadmap artifact json
// missing" — pointing the operator at the very cascade guided just declared
// optional. Seeding is called only when the resolved profile does not require
// roadmap grading; under strict the artifacts must already exist and
// preflightArtifacts still refuses a missing one exactly as before.
func SeedArtifactsIfAbsent() error {
	seeds := []struct{ path, body string }{
		{RoadmapAnalysisFile, seedAnalysisJSON},
		{RoadmapQualityFile, seedQualityJSON},
		{RoadmapAnalysisMarkdown, seedAnalysisMD},
		{RoadmapQualityMarkdown, seedQualityMD},
	}
	for _, seed := range seeds {
		if _, err := os.Stat(seed.path); err == nil {
			continue
		}
		if err := writeAtomic(seed.path, []byte(seed.body)); err != nil {
			return err
		}
	}
	return nil
}

// seedThenPreflight is the single seam both promote branches use, placed AFTER
// every slug and phase check so an invalid request writes nothing at all — the
// "rejected before any write" posture promote otherwise keeps. Seeding before
// validation left four unrequested files behind on a bad slug.
func seedThenPreflight(req PromoteRequest) error {
	if req.SeedArtifacts {
		if err := SeedArtifactsIfAbsent(); err != nil {
			return err
		}
	}
	return preflightArtifacts()
}
