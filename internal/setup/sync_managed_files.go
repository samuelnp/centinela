package setup

import (
	"os"
	"path/filepath"
	"strings"
)

const pluginHeader = "// centinela:managed-version=" + setupDocVersion + " template=.opencode/plugins/centinela.js"
const agentsHeader = "<!-- centinela:managed-version=" + setupDocVersion + " template=AGENTS.md -->"

func planPluginFile() (*SyncItem, error) {
	return planManagedFile(pluginFile, pluginHeader+"\n"+pluginContent, pluginContent, SyncKindPrewriteHook)
}

func planAgentsFile() (*SyncItem, error) {
	return planManagedFile(agentsFile, agentsHeader+"\n"+agentsContent, agentsContent, SyncAgents)
}

func planManagedFile(path, target, legacy string, kind SyncKind) (*SyncItem, error) {
	cur, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &SyncItem{Kind: kind, Path: path, Action: SyncCreate}, nil
	}
	if err != nil {
		return nil, err
	}
	s := string(cur)
	if s == target {
		return nil, nil
	}
	if strings.HasPrefix(s, "// centinela:managed-version=") ||
		strings.HasPrefix(s, "<!-- centinela:managed-version=") ||
		strings.HasPrefix(s, "# centinela:managed-version=") {
		return &SyncItem{Kind: kind, Path: path, Action: SyncUpdate}, nil
	}
	if s == legacy {
		return &SyncItem{Kind: kind, Path: path, Action: SyncUpdate}, nil
	}
	return &SyncItem{Kind: kind, Path: path, Action: SyncManualReview, Reason: "unmanaged custom content"}, nil
}

func writeManagedPlugin(path string) error {
	return writeManaged(path, pluginHeader+"\n"+pluginContent)
}

func writeManagedAgents(path string) error {
	return writeManaged(path, agentsHeader+"\n"+agentsContent)
}

func writeManaged(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0644)
}
