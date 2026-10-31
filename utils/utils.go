package utils

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"unicode"
)

func TruncateText(str string, max int) string {
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

func BrowserOpen(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Println(err)
	}
}

func GetNewLines(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("\n", n)
}

func MaxInt(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
