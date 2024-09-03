package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrnxf/gh-eco/ui"
)

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, _ := tea.LogToFile("debug.log", "")
		defer f.Close()
	}

	p := tea.NewProgram(ui.New(), tea.WithAltScreen(), tea.WithMouseAllMotion())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}
