package evidence

// changelogBody renders the .workflow/<feature>-changelog.md stub the docs
// step requires for internal features. The first non-blank line is the entry;
// replace the FILL slots with a one-line conventional-commit-shaped summary.
func changelogBody(_ string) []byte {
	return []byte("- " + FillSlot("type") + ": " + FillSlot("one-line summary of the change") + "\n")
}
