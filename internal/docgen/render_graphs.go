package docgen

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

func mermaidRoadmap(nodes []RoadmapNode) string {
	b := &strings.Builder{}
	b.WriteString("flowchart LR\n")
	if len(nodes) == 0 {
		b.WriteString("  n0[No features detected]\n")
		return b.String()
	}
	ids := map[string]string{}
	for i, n := range nodes {
		id := fmt.Sprintf("f%d", i)
		ids[n.Name] = id
		b.WriteString("  " + id + "[\"" + esc(n.Name) + "\"]\n")
	}
	for _, n := range nodes {
		for _, dep := range n.DependsOn {
			if ids[dep] == "" {
				id := cleanID(dep)
				ids[dep] = id
				b.WriteString("  " + id + "[\"" + esc(dep) + "\"]\n")
			}
			b.WriteString("  " + ids[dep] + " --> " + ids[n.Name] + "\n")
		}
	}
	return b.String()
}

func mermaidSpecs(specs []string) string {
	b := &strings.Builder{}
	b.WriteString("flowchart TD\n  root[\"Project Specs\"]\n")
	for i, s := range specs {
		id := fmt.Sprintf("s%d", i)
		name := strings.TrimSuffix(filepath.Base(s), filepath.Ext(s))
		b.WriteString("  " + id + "[\"" + esc(name) + "\"]\n")
		b.WriteString("  root --> " + id + "\n")
	}
	if len(specs) == 0 {
		b.WriteString("  root --> empty[\"No specs found\"]\n")
	}
	return b.String()
}

func cleanID(s string) string {
	b := &strings.Builder{}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "node"
	}
	return b.String()
}
