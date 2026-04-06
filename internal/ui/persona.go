package ui

func personaFace(t tone) string {
	switch t {
	case toneSuccess:
		return "^_^"
	case toneWarn:
		return "-_-"
	case toneError:
		return "ò_ó"
	default:
		return "o_o"
	}
}

func personaLabel(t tone) string {
	return " CENTINELA says " + personaFace(t) + " "
}
