// Package docsctx assembles the curated feature-scale inputs the
// documentation specialist reads in the docs step: feature brief, plan,
// spec scenarios, and the optional changelog draft.
package docsctx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Section is one embedded input file: its source path and verbatim body.
// Present is false when the file does not exist on disk.
type Section struct {
	Source  string
	Body    string
	Present bool
}

// Context carries the four docs-step inputs for one feature.
type Context struct {
	Feature   string
	Brief     Section
	Plan      Section
	Spec      Section
	Changelog Section
}

// Load resolves the docs-step input files for a feature. Brief, plan and
// spec are required — ALL missing ones are aggregated into a single error.
// The changelog draft is optional and simply marked absent.
func Load(feature string) (Context, error) {
	ctx := Context{
		Feature:   feature,
		Brief:     readSection(filepath.Join("docs", "features", feature+".md")),
		Plan:      readSection(filepath.Join("docs", "plans", feature+".md")),
		Spec:      readSection(filepath.Join("specs", feature+".feature")),
		Changelog: readSection(filepath.Join(".workflow", feature+"-changelog.md")),
	}
	missing := []string{}
	for _, s := range []Section{ctx.Brief, ctx.Plan, ctx.Spec} {
		if !s.Present {
			missing = append(missing, s.Source)
		}
	}
	if len(missing) > 0 {
		return Context{}, fmt.Errorf("docs context inputs missing for %q: %s (brief, plan and spec are written in the plan step)", feature, strings.Join(missing, ", "))
	}
	return ctx, nil
}

// readSection reads one input file verbatim; a missing or unreadable file
// yields an absent Section that still names its expected source path.
func readSection(path string) Section {
	src := filepath.ToSlash(path)
	body, err := os.ReadFile(path)
	if err != nil {
		return Section{Source: src}
	}
	return Section{Source: src, Body: string(body), Present: true}
}
