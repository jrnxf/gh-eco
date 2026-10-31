package github

import (
	"encoding/base64"
	"strings"
	"testing"
)

type decodeReadmeContentTest struct {
	name      string
	content   string
	encoding  string
	expected  string
	expectErr bool
}

func Test_DecodeReadmeContent(t *testing.T) {
	happy := base64.StdEncoding.EncodeToString([]byte("# Hello\n"))

	// Split a valid base64 encoding with a newline mid-string, mirroring
	// GitHub's wrapped output.
	full := base64.StdEncoding.EncodeToString([]byte("# Hello world readme\n"))
	mid := len(full) / 2
	wrapped := full[:mid] + "\n" + full[mid:]

	tests := []decodeReadmeContentTest{
		{
			name:     "happy path",
			content:  happy,
			encoding: "base64",
			expected: "# Hello\n",
		},
		{
			name:     "newline-embedded base64",
			content:  wrapped,
			encoding: "base64",
			expected: "# Hello world readme\n",
		},
		{
			name:      "wrong encoding",
			content:   happy,
			encoding:  "utf-8",
			expectErr: true,
		},
		{
			name:      "invalid base64",
			content:   "!!!",
			encoding:  "base64",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := decodeReadmeContent(test.content, test.encoding)
			if test.expectErr {
				if err == nil {
					t.Errorf("expected an error, got nil (output %q)", output)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if output != test.expected {
				t.Errorf("output %q not equal to expected %q", output, test.expected)
			}
		})
	}
}

func Test_DecodeReadmeContent_WrongEncodingMentionsEncoding(t *testing.T) {
	_, err := decodeReadmeContent("", "utf-8")
	if err == nil {
		t.Fatal("expected an error for non-base64 encoding")
	}
	if got := err.Error(); !strings.Contains(got, "utf-8") {
		t.Errorf("error %q does not mention the encoding", got)
	}
}
