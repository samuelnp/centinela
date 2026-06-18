package analyze

import "sort"

// extensionLanguage maps a lowercased file extension (with leading dot) to a
// language display name. Adding a language is a table edit, not new control
// flow. Unknown extensions are ignored (not counted as a language).
var extensionLanguage = map[string]string{
	".go":    "Go",
	".js":    "JavaScript",
	".jsx":   "JavaScript",
	".mjs":   "JavaScript",
	".cjs":   "JavaScript",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".rb":    "Ruby",
	".rs":    "Rust",
	".py":    "Python",
	".java":  "Java",
	".kt":    "Kotlin",
	".swift": "Swift",
	".c":     "C",
	".h":     "C",
	".cc":    "C++",
	".cpp":   "C++",
	".hpp":   "C++",
	".cs":    "C#",
	".php":   "PHP",
	".sh":    "Shell",
	".sql":   "SQL",
}

// detectLanguages converts per-extension file counts into LanguageStats summed
// by language. The result is sorted by file count descending, then name
// ascending (a deterministic tiebreak). primary is the name of the top entry,
// or "" when nothing was counted.
func detectLanguages(extCounts map[string]int) (stats []LanguageStat, primary string) {
	byLang := map[string]int{}
	for ext, n := range extCounts {
		if lang, ok := extensionLanguage[ext]; ok {
			byLang[lang] += n
		}
	}
	for name, count := range byLang {
		stats = append(stats, LanguageStat{Name: name, FileCount: count})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].FileCount != stats[j].FileCount {
			return stats[i].FileCount > stats[j].FileCount
		}
		return stats[i].Name < stats[j].Name
	})
	if len(stats) > 0 {
		primary = stats[0].Name
	}
	return stats, primary
}
