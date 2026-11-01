package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/ui"
	"github.com/jrnxf/gh-eco/ui/theme"
)

const usage = `usage: gh eco [<login>]

Opens the profile of a GitHub user or organization, or a search box when
no login is given.`

// errHelp asks for the usage text.
var errHelp = errors.New("help requested")

// parseArgs returns the login to open, or "" to start in the search box.
func parseArgs(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	if len(args) > 1 {
		return "", fmt.Errorf("expected at most one login, got %d arguments", len(args))
	}
	arg := args[0]
	switch {
	case arg == "-h" || arg == "--help":
		return "", errHelp
	case strings.HasPrefix(arg, "-"):
		return "", fmt.Errorf("unknown flag %s", arg)
	}
	// a login copied from a mention keeps its @
	login := strings.TrimPrefix(strings.TrimSpace(arg), "@")
	if login == "" {
		return "", errors.New("login is empty")
	}
	return login, nil
}

func main() {
	login, err := parseArgs(os.Args[1:])
	if errors.Is(err, errHelp) {
		fmt.Println(usage)
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to open debug.log:", err)
			os.Exit(1)
		}
		defer f.Close()
	} else {
		log.SetOutput(io.Discard)
	}

	// before anything renders, so adaptive colors match the terminal
	theme.Detect()

	client, err := github.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gh-eco requires GitHub CLI authentication: run `gh auth login` and try again.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(client, login), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
