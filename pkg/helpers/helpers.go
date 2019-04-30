package helpers

import "strings"

// SplitOnFirstDelim - Splits string into two parts at the delim character
func SplitOnFirstDelim(delim rune, s string) (string, string) {
	split := strings.SplitN(s, string(delim), 2)
	if len(split) == 1 {
		return strings.TrimSpace(split[0]), ""
	}
	return strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
}
