package docgen

import "strings"

func mermaidRoadmap(nodes []RoadmapNode) string {
	b := &strings.Builder{}
	b.WriteString("flowchart LR\n")
	for _, n := range nodes {
		if len(n.DependsOn) == 0 {
			b.WriteString("  " + n.Name + "\n")
			continue
		}
		for _, dep := range n.DependsOn {
			b.WriteString("  " + dep + " --> " + n.Name + "\n")
		}
	}
	return b.String()
}

func mermaidEvidence(ev []EvidenceLink) string {
	b := &strings.Builder{}
	b.WriteString("flowchart TD\n")
	for _, e := range ev {
		if e.Handoff == "" {
			continue
		}
		b.WriteString("  " + e.Role + " --> " + e.Handoff + "\n")
	}
	return b.String()
}
