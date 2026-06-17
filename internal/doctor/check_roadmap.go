package doctor

import (
	"bytes"
	"os"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const roadmapMarkdownFile = "ROADMAP.md"

// roadmapCheck diagnoses (a) ROADMAP.md ↔ roadmap.json drift and (b) a
// live-status glyph baked into a phase name (e.g. "✅ Phase 0: Bootstrap"),
// which defeats isBootstrapPhaseName's "phase 0" prefix match. Both repairs are
// safe+idempotent: strip any glyph, then regenerate ROADMAP.md from roadmap.json.
type roadmapCheck struct{}

func (roadmapCheck) Name() string { return "roadmap" }

func (roadmapCheck) Run(Context) Diagnosis {
	d := Diagnosis{Name: "roadmap"}
	rm, err := roadmap.Load()
	if err != nil {
		if os.IsNotExist(err) {
			d.Status = OK
			d.Message = "no roadmap.json — roadmap check not applicable"
			return d
		}
		d.Status = Error
		d.Message = "cannot load " + roadmap.RoadmapFile + ": " + err.Error()
		return d
	}
	glyphs := glyphPhases(rm)
	want := roadmap.RenderMarkdown(rm)
	got, _ := os.ReadFile(roadmapMarkdownFile)
	drifted := !bytes.Equal(want, got)
	if len(glyphs) == 0 && !drifted {
		d.Status = OK
		d.Message = "ROADMAP.md in sync and no phase-name glyphs"
		return d
	}
	d.Status = Error
	d.Message = describeRoadmap(glyphs, drifted)
	for _, g := range glyphs {
		d.Details = append(d.Details, "phase glyph breaks prefix detection: "+g)
	}
	if drifted {
		d.Details = append(d.Details, roadmapMarkdownFile+" is out of sync with roadmap.json")
	}
	d.Repair = &Repair{Safe: true, Idempotent: true, Apply: repairRoadmap}
	return d
}
