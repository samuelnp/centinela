package setup

const aiderConfigFile = ".aider.conf.yml"
const aiderConfigBody = "read: AGENTS.md\n"
const aiderConfigHeader = "# centinela:managed-version=" + setupDocVersion + " template=.aider.conf.yml"

// planAiderConfig plans the managed .aider.conf.yml file via the shared
// managed-marker seam: absent -> create, managed -> update, unmanaged ->
// manual-review (never clobbered).
func planAiderConfig() (*SyncItem, error) {
	return planManagedFile(aiderConfigFile, aiderConfigHeader+"\n"+aiderConfigBody, aiderConfigBody, SyncAiderConfig)
}

func writeManagedAiderConfig(path string) error {
	return writeManaged(path, aiderConfigHeader+"\n"+aiderConfigBody)
}
