package analyze

// Analyze produces a complete Inventory for the repository rooted at root. It is
// diagnostic, not a gate: only an unreadable root is a hard error. Every
// sub-detector (manifests, locales, graph) degrades to a best-effort/empty
// result so the command always emits a valid inventory and exits 0 (Decision
// #3). The returned Inventory has all slices pre-sorted for byte-stable Save.
func Analyze(root string) (Inventory, error) {
	wr, err := walk(root)
	if err != nil {
		return Inventory{}, err
	}
	langs, primary := detectLanguages(wr.extCounts)
	manifests := detectManifests(root)
	inv := Inventory{
		SchemaVersion:   SchemaVersion,
		PrimaryLanguage: primary,
		Languages:       langs,
		Manifests:       manifests,
		Locales:         detectLocales(root),
		Packages:        wr.packages,
		Graph:           buildGraph(manifests),
	}
	return inv, nil
}
