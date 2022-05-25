package main

import (
	"fmt"
	"log"
	"os"

	"github.com/coloradocolby/gh-eco/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	f, err := tea.LogToFile("debug.log", "")
	// clear the file each run

	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}

	log.Println("\n\n\n\n\nPROGRAM START")

	p := tea.NewProgram(ui.New(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
	defer f.Close()

}
