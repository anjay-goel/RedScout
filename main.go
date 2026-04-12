package main

import (
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"redscout/lib"
	"redscout/lib/ui"
)

func main() {
	config := lib.ParseFlags()

	model := ui.NewModel(config)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		log.Fatal(err)
	}
}
