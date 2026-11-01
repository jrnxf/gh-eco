# Records the README demo on live data; ./demo.sh renders it with keycaps.
# Needs a gh login. Never add s or ! here: they star repos publicly.

Output demo.gif

Set Shell bash
Set Width 2560
Set Height 1450
Set FontFamily "JetBrains Mono"
Set FontSize 30
Set Padding 24
Set Framerate 20
Set TypingSpeed 60ms
Set WaitTimeout 30s
Set Theme { "name": "gh-eco", "background": "#1d2025", "foreground": "#d0d0d0", "cursor": "#d0d0d0", "selection": "#3a3f47", "black": "#1d2025", "red": "#ff6b6b", "green": "#3fb950", "yellow": "#e3b341", "blue": "#58a6ff", "magenta": "#bc8cff", "cyan": "#39c5cf", "white": "#d0d0d0", "brightBlack": "#6e7681", "brightRed": "#ff7b72", "brightGreen": "#56d364", "brightYellow": "#f2cc60", "brightBlue": "#79c0ff", "brightMagenta": "#d2a8ff", "brightCyan": "#56d4dd", "brightWhite": "#ffffff" }

Hide
Type "go build -o gh-eco . && clear"
Enter
Wait@3m
Type "GH_ECO_THEME=dark ./gh-eco"
Enter
Sleep 1s
Show

# own profile; network waits happen off camera
Type "jrnxf"
Sleep 200ms
Enter
Sleep 150ms
Hide
Wait+Screen /contributions/
Show
Sleep 700ms

# the gh-eco card, its readme, a little scroll, back
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 500ms
Type "r"
Sleep 150ms
Hide
Wait+Screen /Installation/
Show
Sleep 1.2s
Ctrl+D
Sleep 600ms
Ctrl+D
Sleep 1s
Type "r"
Sleep 600ms

# someone else
Type "/"
Sleep 300ms
Type "charmbracelet"
Sleep 200ms
Enter
Sleep 150ms
Hide
Wait+Screen /charm\.land/
Show
Sleep 800ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 1.6s

# quit off camera, so the loop ends on the app
Hide
Escape
