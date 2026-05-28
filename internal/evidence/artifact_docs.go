package evidence

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// docsSpecialistPair renders both the JSON evidence and its markdown
// narrative for the documentation-specialist role. The JSON is built
// through the same schema/marshal pipeline as `centinela evidence init` so
// the on-disk key order is stable.
func docsSpecialistPair(feature string) ([]string, [][]byte, error) {
	role := orchestration.RoleDocsSpecialist
	skel := Skeleton(feature, role, "0.0.0")
	skel.Meta = nil // template should not pin a CLI version
	skel.GeneratedAt = today() + "T00:00:00Z"
	body, err := skel.MarshalJSON()
	if err != nil {
		return nil, nil, fmt.Errorf("docs template marshal: %w", err)
	}
	md := []byte(fmt.Sprintf(`# Orchestration Evidence: documentation-specialist

- Feature: %s
- Step: docs
- Outcome: _One paragraph summary of the KB pages and project-docs entries you wrote._
- Handoff: complete
`, "`"+feature+"`"))
	paths := []string{
		artifactPath(feature, "documentation-specialist.json"),
		artifactPath(feature, "documentation-specialist.md"),
	}
	return paths, [][]byte{body, md}, nil
}
