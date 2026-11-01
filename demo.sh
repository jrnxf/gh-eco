#!/usr/bin/env bash
# Renders the README demo: vhs records demo.tape, then ffmpeg draws each
# key press as a small keycap so viewers can follow along. Typed words are
# left out, since the search bar already shows them.
#
# Needs vhs, ffmpeg built with drawtext, a gh login, and the JetBrains Mono
# font (vhs's default font has no ⭑ for star counts).
set -euo pipefail
cd "$(dirname "$0")"

vhs demo.tape

# Replay the tape's clock from Show to find when each key lands. vhs waits
# TypingSpeed after each typed character but not after Enter, Escape or
# Ctrl keys, and records only after Show.
filters=$(awk '
	function secs(d) {
		if (d ~ /ms$/) return substr(d, 1, length(d) - 2) / 1000
		if (d ~ /s$/) return substr(d, 1, length(d) - 1)
		if (d ~ /m$/) return substr(d, 1, length(d) - 1) * 60
		return d
	}
	function cap(label) {
		if (shown) { n++; at[n] = t; key[n] = label }
	}
	$1 == "Set" && $2 == "TypingSpeed" { speed = secs($3) }
	$1 == "Show" { shown = 1 }
	$1 == "Hide" { shown = 0 }
	$1 == "Sleep" && shown { t += secs($2) }
	$1 == "Type" {
		match($0, /"[^"]*"/)
		s = substr($0, RSTART + 1, RLENGTH - 2)
		if (length(s) == 1) cap(s)
		if (shown) t += length(s) * speed
	}
	$1 == "Enter" { cap("enter") }
	$1 == "Escape" { cap("esc") }
	$1 ~ /^Ctrl\+/ { cap(tolower($1)) }
	END {
		# each keycap stays up to 0.8 s, and clears just before the next one
		for (i = 1; i <= n; i++) {
			end = at[i] + 0.8
			if (i < n && at[i + 1] - 0.05 < end) end = at[i + 1] - 0.05
			f = f sprintf(",drawtext=font=JetBrains Mono:text=%s:fontsize=30:fontcolor=0xd0d0d0:box=1:boxcolor=0x3a3f47@0.9:boxborderw=14|24:x=w-tw-72:y=36:enable=between(t\\,%.2f\\,%.2f)", key[i], at[i], end)
		}
		print substr(f, 2)
	}
' demo.tape)

ffmpeg -loglevel error -y -i demo.gif \
	-filter_complex "${filters},split[a][b];[a]palettegen=stats_mode=diff[p];[b][p]paletteuse=dither=none" \
	demo.keys.gif
mv demo.keys.gif demo.gif
