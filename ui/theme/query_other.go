//go:build !unix

package theme

func queryBackground() (reply string, ok bool) {
	return "", false
}
