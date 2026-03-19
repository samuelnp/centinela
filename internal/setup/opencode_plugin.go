package setup

import "os"

const pluginFile = ".opencode/plugins/centinela.js"

const pluginContent = `export const CentinelaPlugin = async () => {
  return {
    "tool.execute.before": async (input) => {
      if (!isWriteTool(input.tool)) return
      const filePath = getFilePath(input.args)
      if (!filePath) return

      const payload = JSON.stringify({ tool_input: { filePath } })
      const proc = Bun.spawnSync({
        cmd: ["centinela", "hook", "prewrite"],
        stdin: payload,
        stderr: "pipe",
      })
      if (proc.exitCode === 0) return
      const msg = new TextDecoder().decode(proc.stderr).trim()
      throw new Error(msg || "centinela blocked this write")
    },
  }
}

function isWriteTool(tool) {
  return tool === "write" || tool === "edit" || tool === "patch"
}

function getFilePath(args) {
  if (!args || typeof args !== "object") return ""
  return args.filePath || args.file_path || ""
}
`

// EnsureOpenCodePlugin writes the Centinela OpenCode plugin if missing.
func EnsureOpenCodePlugin() (bool, error) {
	if _, err := os.Stat(pluginFile); err == nil {
		return false, nil
	}
	if err := os.MkdirAll(".opencode/plugins", 0755); err != nil {
		return false, err
	}
	return true, os.WriteFile(pluginFile, []byte(pluginContent), 0644)
}
