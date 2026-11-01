package repo

import "testing"

func Test_TruncateText(t *testing.T) {
	tests := []struct {
		text     string
		max      int
		expected string
	}{
		{"the quick brown fox jumped over the lazy dog", 15, "the quick brown…"},
		{"the quick brown fox jumped over the lazy dog", 14, "the quick…"},
		{"the quick brown fox jumped over the lazy dog", -14, ""},
		{"the quick brown fox jumped over the lazy dog", 0, ""},
		{"the quick brown fox jumped over the lazy dog", 1, "the…"},
		{"abcdefghij", 5, "abcde…"}, // no spaces: hard cut
		{"éééééééééé", 3, "ééé…"},   // multibyte rune safety
		{"short", 10, "short"},      // under max: unchanged
	}

	for _, test := range tests {
		if output := truncateText(test.text, test.max); output != test.expected {
			t.Errorf("truncateText(%q, %d) = %q, want %q", test.text, test.max, output, test.expected)
		}
	}
}
