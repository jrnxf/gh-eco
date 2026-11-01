//go:build unix

package theme

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// replyTimeout bounds the wait for a terminal that never answers.
const replyTimeout = 100 * time.Millisecond

// queryBackground asks the terminal for its background color (OSC 11).
func queryBackground() (r, g, b float64, ok bool) {
	in, out := int(os.Stdin.Fd()), int(os.Stdout.Fd())
	if !term.IsTerminal(in) || !term.IsTerminal(out) {
		return 0, 0, 0, false
	}
	state, err := term.MakeRaw(in)
	if err != nil {
		return 0, 0, 0, false
	}
	defer term.Restore(in, state)

	if _, err := os.Stdout.WriteString("\x1b]11;?\x07"); err != nil {
		return 0, 0, 0, false
	}

	var reply []byte
	deadline := time.Now().Add(replyTimeout)
	buf := make([]byte, 64)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return 0, 0, 0, false
		}
		var fds unix.FdSet
		fds.Set(in)
		tv := unix.NsecToTimeval(left.Nanoseconds())
		n, err := unix.Select(in+1, &fds, nil, nil, &tv)
		if err != nil || n == 0 {
			return 0, 0, 0, false
		}
		k, err := unix.Read(in, buf)
		if err != nil || k == 0 {
			return 0, 0, 0, false
		}
		reply = append(reply, buf[:k]...)
		if last := reply[len(reply)-1]; last == '\a' || (last == '\\' && len(reply) > 1 && reply[len(reply)-2] == '\x1b') {
			return parseOSC11(string(reply))
		}
	}
}
