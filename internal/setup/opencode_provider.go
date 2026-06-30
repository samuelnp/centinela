package setup

// LocalProvider is the setup-local value driving the managed OpenCode provider
// block. It is mapped from config.LocalConfig in cmd/ so internal/setup keeps
// importing nothing internal. A nil *LocalProvider means "no local block" and
// every provider step no-ops, preserving the zero-config managed output.
type LocalProvider struct {
	Provider  string
	Endpoint  string
	Model     string
	APIKeyEnv string
}

// openCompatNPM is the npm package both local provider kinds use; OpenCode
// drives any OpenAI-compatible endpoint (Ollama, llama.cpp, vLLM, LM Studio)
// through it.
const openCompatNPM = "@ai-sdk/openai-compatible"

// buildLocalProvider returns the provider key (the provider name) and the
// managed provider block: npm @ai-sdk/openai-compatible, options.baseURL from
// the endpoint, an options.apiKey env reference for openai-compatible when set,
// and the declared model under models.
func buildLocalProvider(lp LocalProvider) (string, map[string]any) {
	options := map[string]any{"baseURL": lp.Endpoint}
	if lp.Provider == "openai-compatible" && lp.APIKeyEnv != "" {
		options["apiKey"] = "{env:" + lp.APIKeyEnv + "}"
	}
	block := map[string]any{
		"npm":     openCompatNPM,
		"options": options,
		"models":  map[string]any{lp.Model: map[string]any{}},
	}
	return lp.Provider, block
}
