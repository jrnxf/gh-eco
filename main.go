package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/coloradocolby/gh-eco/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	f, err := tea.LogToFile("debug.log", "")
	// clear the file each run
	exec.Command("/bin/bash", "-c", "echo > debug.log").Run()

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
