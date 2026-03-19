package gates

import "fmt"

// flatKeys flattens a nested JSON map to dot-separated keys.
func flatKeys(m map[string]interface{}, prefix string) map[string]bool {
	out := map[string]bool{}
	for k, v := range m {
		full := k
		if prefix != "" {
			full = prefix + "." + k
		}
		if nested, ok := v.(map[string]interface{}); ok {
			for sub := range flatKeys(nested, full) {
				out[sub] = true
			}
		} else {
			out[full] = true
		}
	}
	return out
}

func compareKeysets(keys map[string]map[string]bool, locales []string) Result {
	ref := locales[0]
	var missing []string

	for _, locale := range locales[1:] {
		for k := range keys[ref] {
			if !keys[locale][k] {
				missing = append(missing, fmt.Sprintf("[%s] missing key: %s", locale, k))
			}
		}
		for k := range keys[locale] {
			if !keys[ref][k] {
				missing = append(missing, fmt.Sprintf("[%s] extra key not in %s: %s", locale, ref, k))
			}
		}
	}

	if len(missing) == 0 {
		return Result{Name: "G11: i18n", Status: Pass, Message: "All locales have identical keys."}
	}
	return Result{Name: "G11: i18n", Status: Fail, Message: "Translation keys out of sync.", Details: missing}
}
