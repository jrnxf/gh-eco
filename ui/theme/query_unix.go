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
const replyTimeout = time.Second

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
		if left <= 0 || !readInto(in, &buf, chunk, left) {
			break
		}
	}
	for readInto(in, &buf, chunk, 0) {
	}
	return buf.String(), true
}

// readInto waits up to wait for fd to become readable and appends what
// it reads to buf. It reports whether any bytes were read.
func readInto(fd int, buf *strings.Builder, chunk []byte, wait time.Duration) bool {
	var fds unix.FdSet
	fds.Set(fd)
	tv := unix.NsecToTimeval(wait.Nanoseconds())
	if n, err := unix.Select(fd+1, &fds, nil, nil, &tv); err != nil || n == 0 {
		return false
	}
	k, err := unix.Read(fd, chunk)
	if err != nil || k == 0 {
		return false
	}
	buf.Write(chunk[:k])
	return true
}

// answeredDA1 reports whether s holds a DA1 reply, "\x1b[?...c".
func answeredDA1(s string) bool {
	i := strings.Index(s, "\x1b[?")
	return i >= 0 && strings.Contains(s[i:], "c")
}
