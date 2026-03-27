package setup

import "encoding/json"

type statusLineConfig struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func ensureStatusLine(rawSettings map[string]json.RawMessage) bool {
	if _, ok := rawSettings["statusLine"]; ok {
		return false
	}
	v := statusLineConfig{Type: "command", Command: "centinela hook statusline"}
	rawSettings["statusLine"], _ = json.Marshal(v)
	return true
}
