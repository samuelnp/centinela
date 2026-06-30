package setup

import "testing"

// buildLocalProvider: an ollama block uses the openai-compatible npm package, sets
// options.baseURL from the endpoint, lists the model, and never sets an apiKey.
func TestBuildLocalProviderOllama(t *testing.T) {
	key, block := buildLocalProvider(LocalProvider{Provider: "ollama", Endpoint: "http://localhost:11434/v1", Model: "qwen2.5-coder"})
	if key != "ollama" {
		t.Fatalf("key = %q, want ollama", key)
	}
	if block["npm"] != openCompatNPM {
		t.Fatalf("npm = %v, want %s", block["npm"], openCompatNPM)
	}
	opts := block["options"].(map[string]any)
	if opts["baseURL"] != "http://localhost:11434/v1" {
		t.Fatalf("baseURL = %v", opts["baseURL"])
	}
	if _, ok := opts["apiKey"]; ok {
		t.Fatal("ollama must not set apiKey")
	}
	if _, ok := block["models"].(map[string]any)["qwen2.5-coder"]; !ok {
		t.Fatalf("models missing the declared model: %v", block["models"])
	}
}

// buildLocalProvider: openai-compatible with api_key_env writes an {env:NAME}
// apiKey reference; without api_key_env it omits the apiKey key entirely.
func TestBuildLocalProviderOpenAICompatible(t *testing.T) {
	key, block := buildLocalProvider(LocalProvider{Provider: "openai-compatible", Endpoint: "http://localhost:8000/v1", Model: "llama-3.1-8b", APIKeyEnv: "LOCAL_API_KEY"})
	if key != "openai-compatible" {
		t.Fatalf("key = %q", key)
	}
	opts := block["options"].(map[string]any)
	if opts["baseURL"] != "http://localhost:8000/v1" {
		t.Fatalf("baseURL = %v", opts["baseURL"])
	}
	if opts["apiKey"] != "{env:LOCAL_API_KEY}" {
		t.Fatalf("apiKey ref = %v", opts["apiKey"])
	}

	_, bare := buildLocalProvider(LocalProvider{Provider: "openai-compatible", Endpoint: "http://x/v1", Model: "m"})
	if _, ok := bare["options"].(map[string]any)["apiKey"]; ok {
		t.Fatal("missing api_key_env must omit apiKey")
	}
}
