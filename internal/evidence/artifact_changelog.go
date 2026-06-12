package evidence

// changelogBody renders the .workflow/<feature>-changelog.md stub the docs
// step requires for internal features. The first non-blank line is the entry;
// replace the placeholder with a one-line summary of the change.
func changelogBody(_ string) []byte {
	return []byte("- <type>: <one-line summary of the change>\n")
}
