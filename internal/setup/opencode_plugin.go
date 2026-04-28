package setup

import "os"

const pluginFile = ".opencode/plugins/centinela.js"

const pluginContent = `export const CentinelaPlugin = async () => {
  return {
    "tool.execute.before": async (input) => {
      if (!isWriteTool(normalizeTool(input))) return
      const filePath = getFilePath(input)
      if (!filePath) return
      const payload = JSON.stringify({ tool_input: { filePath } })
      runHook("prewrite", payload, true)
    },
    "tool.execute.after": async (input) => {
      if (!isWriteTool(normalizeTool(input))) return
      runHook("postwrite", "", false)
    },
    "tui.prompt.append": async (_input, output) => {
      const promptPayload = typeof _input === "string" ? _input : JSON.stringify(_input || {})
      prependContext(output, joinText(
        runHook("setup", "", false),
        runHook("migrate", "", false),
      ))
      appendContext(output, runHook("autostart", promptPayload, false))
      appendContext(output, runHook("orchestration", "", false))
      appendContext(output, runHook("plan-advisor", "", false))
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
function joinText(...parts) {
  return parts.filter(Boolean).join("\n\n")
}
function isWriteTool(tool) {
  tool = String(tool || "").toLowerCase()
  return tool === "write" || tool === "edit" || tool === "patch"
}
function normalizeTool(input) {
  if (!input || typeof input !== "object") return ""
  return input.tool || input.toolName || input.name || ""
}
function getFilePath(input) {
  if (!input || typeof input !== "object") return ""
  const args = input.args && typeof input.args === "object" ? input.args : {}
  const nested = args.input && typeof args.input === "object" ? args.input : {}
  return (
    args.filePath ||
    args.file_path ||
    args.path ||
    args.filename ||
    args.file ||
    nested.filePath ||
    nested.file_path ||
    nested.path ||
    ""
  )
}
function prependContext(output, text) {
  writeContext(output, text, true)
}
function appendContext(output, text) {
  writeContext(output, text, false)
}
function writeContext(output, text, front) {
  if (!text || !output || typeof output !== "object") return
  if (typeof output.prompt === "string") {
    output.prompt = front ? text + "\n\n" + output.prompt : output.prompt + "\n\n" + text
    return
  }
  if (Array.isArray(output.context)) {
    front ? output.context.unshift(text) : output.context.push(text)
  }
}
`

func EnsureOpenCodePlugin() (bool, error) {
	if _, err := os.Stat(pluginFile); err == nil {
		return false, nil
	}
	if err := os.MkdirAll(".opencode/plugins", 0755); err != nil {
		return false, err
	}
	return true, os.WriteFile(pluginFile, []byte(pluginContent), 0644)
}
