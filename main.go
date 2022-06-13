package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thatvegandev/gh-eco/ui"
)

func main() {
	f, _ := tea.LogToFile("debug.log", "")

	p := tea.NewProgram(ui.New(), tea.WithAltScreen(), tea.WithMouseAllMotion())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}

	defer f.Close()
}
