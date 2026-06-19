package synthesize

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/analyze"
)

// renderTechStack fills the Tech Stack table from the inventory's primary
// language, first detected framework, and first non-empty build/test signal.
func renderTechStack(inv analyze.Inventory) string {
	framework, build, test := "<!-- TODO -->", "<!-- TODO -->", "<!-- TODO -->"
	for _, m := range inv.Manifests {
		if framework == "<!-- TODO -->" && m.Framework != "" {
			framework = m.Framework
		}
		if build == "<!-- TODO -->" && m.Build != "" {
			build = "`" + m.Build + "`"
		}
		if test == "<!-- TODO -->" && m.Test != "" {
			test = "`" + m.Test + "`"
		}
	}
	lang := orTODO(inv.PrimaryLanguage)
	i18n := fmt.Sprintf("%d locale(s)", len(inv.Locales))
	return "## Tech Stack\n\n| Concern | Technology |\n|---------|------------|\n" +
		fmt.Sprintf("| Framework | %s |\n| Language | %s |\n| Build | %s |\n| Tests | %s |\n| i18n | %s |",
			framework, lang, build, test, i18n)
}

// renderFolder renders the inventory's package layout as a fenced tree, or a
// TODO stub when no packages were detected.
func renderFolder(inv analyze.Inventory) string {
	if len(inv.Packages) == 0 {
		return "## Folder Structure\n\n```\n<!-- TODO: no packages detected -->\n```"
	}
	var b strings.Builder
	for _, p := range inv.Packages {
		b.WriteString(p + "\n")
	}
	return "## Folder Structure\n\n```\n" + strings.TrimRight(b.String(), "\n") + "\n```"
}

// renderLocales renders the Locales table from the detected locale codes,
// defaulting to en-only when none were found.
func renderLocales(inv analyze.Inventory) string {
	locales := inv.Locales
	if len(locales) == 0 {
		locales = []string{"en"}
	}
	var rows strings.Builder
	for _, code := range locales {
		rows.WriteString(fmt.Sprintf("| `%s` | <!-- language --> |\n", code))
	}
	return "## Locales\n\n| Code | Language |\n|------|----------|\n" + strings.TrimRight(rows.String(), "\n")
}

func orTODO(s string) string {
	if strings.TrimSpace(s) == "" {
		return "<!-- TODO -->"
	}
	return s
}
