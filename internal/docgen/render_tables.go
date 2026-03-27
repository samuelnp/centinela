package docgen

import "strings"

func statesTable(states []FeatureState) string {
	b := &strings.Builder{}
	b.WriteString("<table><tr><th>Feature</th><th>Current Step</th><th>Status</th></tr>")
	for _, s := range states {
		b.WriteString("<tr><td><code>" + esc(s.Feature) + "</code></td><td>" + esc(s.Step) + "</td><td>" + esc(s.Status) + "</td></tr>")
	}
	b.WriteString("</table>")
	return b.String()
}

func evidenceTable(items []EvidenceLink) string {
	b := &strings.Builder{}
	b.WriteString("<table><tr><th>Role</th><th>Feature</th><th>Step</th><th>Code refs</th></tr>")
	for _, it := range items {
		b.WriteString("<tr><td>" + esc(it.Role) + "</td><td><code>" + esc(it.Feature) + "</code></td><td>" + esc(it.Step) + "</td><td>")
		for i, out := range it.Outputs {
			if i > 0 {
				b.WriteString("<br>")
			}
			b.WriteString("<code>" + esc(out) + "</code>")
		}
		b.WriteString("</td></tr>")
	}
	b.WriteString("</table>")
	return b.String()
}
