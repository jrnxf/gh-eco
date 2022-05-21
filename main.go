package main

import (
	"fmt"
	"log"
	"os"

	"github.com/coloradocolby/ghx/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(), tea.WithAltScreen()) // tea.WithMouseCellMotion()

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
	defer f.Close()

}
