package synthesize

import (
	"strings"

	"github.com/samuelnp/centinela/internal/analyze"
)

// namingProfile is the language-specific test-file and identifier conventions
// surfaced in the Naming Conventions section.
type namingProfile struct {
	testSuffix string
	identifier string
}

// namingByLang maps a lowercased primary language to its conventions. Unknown
// languages fall back to a generic row set.
var namingByLang = map[string]namingProfile{
	"go":         {"_test.go", "PascalCase exported, camelCase unexported"},
	"typescript": {".test.ts", "PascalCase types, camelCase values"},
	"javascript": {".test.js", "PascalCase classes, camelCase values"},
	"ruby":       {"_spec.rb", "snake_case files, CamelCase classes"},
	"python":     {"_test.py", "snake_case modules, PascalCase classes"},
	"rust":       {" (in-file #[test])", "snake_case modules, CamelCase types"},
}

// renderNaming fills the Naming Conventions table from the primary language.
func renderNaming(inv analyze.Inventory) string {
	np, ok := namingByLang[strings.ToLower(inv.PrimaryLanguage)]
	if !ok {
		np = namingProfile{testSuffix: "<!-- TODO -->", identifier: "<!-- TODO: confirm -->"}
	}
	return "## Naming Conventions\n\n| Aspect | Convention |\n|--------|------------|\n" +
		"| Identifiers | " + np.identifier + " |\n" +
		"| Test file | mirrors the file under test + `" + np.testSuffix + "` |\n" +
		"| Spec | kebab-case + `.feature` |"
}
