//go:build !unix

package theme

func queryBackground() (r, g, b float64, ok bool) {
	return 0, 0, 0, false
}
