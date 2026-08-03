package main

import "github.com/samuelnp/centinela/internal/roadmap"

// evaluatedArtifact reports whether a grading artifact pair is present AND was
// produced by a real evaluator pass. A `roadmap promote` under guided seeds the
// pair with "provisional": true so work can start; those files must not satisfy
// a STRICT rung whose whole purpose is to demand the evaluation, or `hook setup`
// would wave through the exact tree on which `start` refuses with the provisional
// message. ValidateAnalysis/ValidateQuality already refuse it; this is the same
// question asked at the cascade.
func evaluatedArtifact(markdown, jsonPath string) bool {
	return exists(markdown) && exists(jsonPath) && !roadmap.ArtifactIsProvisional(jsonPath)
}
