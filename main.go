package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/ui"
)

func main() {
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

	client, err := github.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gh-eco requires GitHub CLI authentication: run `gh auth login` and try again.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(client), tea.WithAltScreen(), tea.WithMouseAllMotion())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
