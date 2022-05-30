package utils

import "testing"

type getNewLinesTest struct {
	arg1     int
	expected string
}

func Test_GetNewLines(t *testing.T) {
	var tests = []getNewLinesTest{
		{4, "\n\n\n\n"},
		{0, ""},
		{-14, ""},
	}

	for _, test := range tests {
		if output := GetNewLines(test.arg1); output != test.expected {
			t.Errorf("Output %q not equal to expected %q", output, test.expected)
		}
	}
}

type truncateTextTest struct {
	arg1     string
	arg2     int
	expected string
}

func Test_TruncateText(t *testing.T) {
	var tests = []truncateTextTest{
		{"the quick brown fox jumped over the lazy dog", 15, "the quick brown..."},
		{"the quick brown fox jumped over the lazy dog", 14, "the quick..."},
		{"the quick brown fox jumped over the lazy dog", -14, ""},
		{"the quick brown fox jumped over the lazy dog", 0, ""},
		{"the quick brown fox jumped over the lazy dog", 1, "the..."},
	}

	for _, test := range tests {
		if output := TruncateText(test.arg1, test.arg2); output != test.expected {
			t.Errorf("Output %q not equal to expected %q", output, test.expected)
		}
	}
}

type maxIntTest struct {
	arg1     int
	arg2     int
	expected int
}

func Test_MaxInt(t *testing.T) {
	var tests = []maxIntTest{
		{1, 3, 3},
		{-1, 99, 99},
		{54, 54, 54},
		{312, 4, 312},
	}

	for _, test := range tests {
		if output := MaxInt(test.arg1, test.arg2); output != test.expected {
			t.Errorf("Output %q not equal to expected %q", output, test.expected)
		}
	}
}
