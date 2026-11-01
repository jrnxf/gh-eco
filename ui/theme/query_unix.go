//go:build unix

package theme

import (
	"os"
	"strings"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// replyTimeout bounds the wait for a terminal that answers neither query.
const replyTimeout = 150 * time.Millisecond

// queryBackground asks the terminal for its background color (OSC 11),
// followed by a device attributes request (DA1) that every terminal
// answers. The DA1 reply ends the wait at once on terminals that ignore
// OSC 11. ok is false when stdin or stdout is not a terminal.
func queryBackground() (reply string, ok bool) {
	in, out := int(os.Stdin.Fd()), int(os.Stdout.Fd())
	if !term.IsTerminal(in) || !term.IsTerminal(out) {
		return "", false
	}
	state, err := term.MakeRaw(in)
	if err != nil {
		return "", false
	}
	defer term.Restore(in, state)

	if _, err := os.Stdout.WriteString("\x1b]11;?\x07\x1b[c"); err != nil {
		return "", true
	}

	var buf strings.Builder
	chunk := make([]byte, 128)
	deadline := time.Now().Add(replyTimeout)
	for !answeredDA1(buf.String()) {
		left := time.Until(deadline)
		if left <= 0 {
			break
		}
		var fds unix.FdSet
		fds.Set(in)
		tv := unix.NsecToTimeval(left.Nanoseconds())
		if n, err := unix.Select(in+1, &fds, nil, nil, &tv); err != nil || n == 0 {
			break
		}
		k, err := unix.Read(in, chunk)
		if err != nil || k == 0 {
			break
		}
		buf.Write(chunk[:k])
	}
	return buf.String(), true
}

// answeredDA1 reports whether s holds a DA1 reply, "\x1b[?...c".
func answeredDA1(s string) bool {
	i := strings.Index(s, "\x1b[?")
	return i >= 0 && strings.Contains(s[i:], "c")
}
