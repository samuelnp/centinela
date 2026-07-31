package docstring

import (
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"sort"
	"strings"
)

// goScanner is the built-in Go implementation of Scanner.
type goScanner struct{}

func init() { Register(GoLang, goScanner{}) }

// Scan parses every in-scope file in files and reports its undocumented
// exported identifiers. A file that cannot be parsed becomes a ParseError, not
// a silent skip; a file that no longer exists on disk is skipped, because a
// scope may name a path that was renamed away.
func (goScanner) Scan(files []string, opts Options) (Report, error) {
	var rep Report
	fset := token.NewFileSet()
	for _, path := range selected(files, opts) {
		src, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			rep.ParseErrors = append(rep.ParseErrors, ParseError{Path: path, Message: firstLine(err.Error())})
			continue
		}
		f, err := parser.ParseFile(fset, path, src, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			rep.ParseErrors = append(rep.ParseErrors, ParseError{Path: path, Message: firstLine(err.Error())})
			continue
		}
		if isGenerated(f) {
			continue
		}
		rep.Files++
		(&collector{path: path, fset: fset, opts: opts, rep: &rep}).file(f)
	}
	sortReport(&rep)
	return rep, nil
}

// firstLine trims a multi-line parser error down to its first line.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

// sortReport orders violations and exemptions by path then line so gate
// details are deterministic across runs and platforms.
func sortReport(r *Report) {
	sort.SliceStable(r.Violations, func(i, j int) bool {
		if r.Violations[i].Path != r.Violations[j].Path {
			return r.Violations[i].Path < r.Violations[j].Path
		}
		return r.Violations[i].Line < r.Violations[j].Line
	})
	sort.SliceStable(r.Exemptions, func(i, j int) bool {
		if r.Exemptions[i].Path != r.Exemptions[j].Path {
			return r.Exemptions[i].Path < r.Exemptions[j].Path
		}
		return r.Exemptions[i].Line < r.Exemptions[j].Line
	})
}
