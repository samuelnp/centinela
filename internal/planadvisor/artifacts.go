package planadvisor

import "fmt"

type artifacts struct{ Brief, Plan, Spec, Edge string }

func loadArtifacts(feature string) artifacts {
	return artifacts{
		Brief: readText(fmt.Sprintf("docs/features/%s.md", feature)),
		Plan:  readText(fmt.Sprintf("docs/plans/%s.md", feature)),
		Spec:  readText(fmt.Sprintf("specs/%s.feature", feature)),
		Edge:  readText(fmt.Sprintf(".workflow/%s-edge-cases.md", feature)),
	}
}
