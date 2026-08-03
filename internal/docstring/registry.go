package docstring

// Scanner inspects an explicit set of files for exported identifiers missing a
// doc comment. It is the language seam: one implementation per language,
// resolved by key. v1 registers Go only.
type Scanner interface {
	// Scan inspects the given files under opts and returns what it found.
	// Files out of scope are the scanner's own concern to skip.
	Scan(files []string, opts Options) (Report, error)
}

// GoLang is the registry key of the built-in Go scanner.
const GoLang = "go"

var scanners = map[string]Scanner{}

// Register binds a Scanner to a language key, replacing any prior binding. It
// mirrors internal/importgraph's provider seam, this repo's precedent for a
// language-pluggable gate.
func Register(lang string, s Scanner) { scanners[lang] = s }

// Unregister drops a language binding. It exists so a caller can prove the
// gate degrades honestly when no scanner is available for a language.
func Unregister(lang string) { delete(scanners, lang) }

// For returns the Scanner registered for a language key.
func For(lang string) (Scanner, bool) {
	s, ok := scanners[lang]
	return s, ok
}
