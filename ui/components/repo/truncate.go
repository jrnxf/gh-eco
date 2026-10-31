package repo

import "unicode"

// truncateText shortens str to at most max runes plus "...", cutting at
// the last space before the limit when there is one.
func truncateText(str string, max int) string {
	if max <= 0 {
		return ""
	}

	lastSpaceIdx := -1
	count := 0
	hardCutIdx := -1 // byte index of the first rune past max; -1 until exceeded
	for i, r := range str {
		if unicode.IsSpace(r) {
			if hardCutIdx != -1 {
				// already past max with no earlier space: cut at this one
				return str[:i] + "..."
			}
			lastSpaceIdx = i
		}
		count++
		if count > max && hardCutIdx == -1 {
			if lastSpaceIdx != -1 {
				return str[:lastSpaceIdx] + "..."
			}
			hardCutIdx = i
		}
	}
	if hardCutIdx != -1 {
		// longer than max with no spaces at all: hard cut at max runes
		return str[:hardCutIdx] + "..."
	}
	return str
}
