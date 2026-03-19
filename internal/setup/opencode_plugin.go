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
      runHook("prewrite", payload, true)
    },

    "tool.execute.after": async (input) => {
      if (!isWriteTool(input.tool)) return
      runHook("postwrite", "", false)
    },

    "tui.prompt.append": async (_input, output) => {
      appendContext(output, runHook("setup", "", false))
      appendContext(output, runHook("context", "", false))
    },
  }
}

function runHook(name, payload, blocking) {
  const proc = Bun.spawnSync({
    cmd: ["centinela", "hook", name],
    stdin: payload,
    stdout: "pipe",
    stderr: "pipe",
  })
  const out = new TextDecoder().decode(proc.stdout).trim()
  const err = new TextDecoder().decode(proc.stderr).trim()
  if (blocking && proc.exitCode !== 0) {
    throw new Error(err || out || "centinela blocked this write")
  }
  return out
}

function isWriteTool(tool) {
  return tool === "write" || tool === "edit" || tool === "patch"
}

function getFilePath(args) {
  if (!args || typeof args !== "object") return ""
  return args.filePath || args.file_path || ""
}

function appendContext(output, text) {
  if (!text || !output || typeof output !== "object") return
  if (typeof output.prompt === "string") {
    output.prompt += "\n\n" + text
    return
  }
  if (Array.isArray(output.context)) {
    output.context.push(text)
  }
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
