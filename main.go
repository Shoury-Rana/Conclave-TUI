package main

import (
	"flag"
	"fmt"
	"os"

	"Conclave-TUI/internal/api"
	"Conclave-TUI/internal/config"
	"Conclave-TUI/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	// Optional CLI flag to override backend address (e.g. -url http://localhost:8000)
	backendURLFlag := flag.String("url", "", "Conclave backend root URL (e.g. http://127.0.0.1:8000)")
	flag.Parse()

	// 1. Load user configuration and stored session tokens (~/.conclave/config.json)
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// CLI flag takes highest precedence over config file
	if *backendURLFlag != "" {
		cfg.BaseURL = *backendURLFlag
	}

	// 2. Initialize the API client
	client := api.NewClient(cfg.BaseURL)

	// 3. Create the root Bubble Tea application model
	app := tui.NewAppModel(cfg, client)

	// 4. Start the Bubble Tea program
	program := tea.NewProgram(app)
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Conclave TUI: %v\n", err)
		os.Exit(1)
	}
}
