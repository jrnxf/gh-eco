package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coloradocolby/gh-eco/ui"
)

func main() {

	f, _ := tea.LogToFile("debug.log", "")

	log.Printf("\n\n\nGH-ECO")

	p := tea.NewProgram(ui.New(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}

	defer f.Close()

}
