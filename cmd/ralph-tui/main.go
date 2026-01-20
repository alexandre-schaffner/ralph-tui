package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alex/ralph-tui/src/lib/process"
	"github.com/alex/ralph-tui/src/lib/state"
	"github.com/alex/ralph-tui/src/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse CLI flags
	mode := flag.String("mode", "build", "Loop mode: build, plan, plan-work")
	maxIter := flag.Int("max", 0, "Max iterations (0 = unlimited)")
	workDesc := flag.String("work", "", "Work description for plan-work mode")
	scriptPath := flag.String("script", "./loop.sh", "Path to loop.sh script")
	flag.Parse()

	// Guard: Validate max iterations is non-negative
	if *maxIter < 0 {
		fmt.Fprintln(os.Stderr, "Error: --max must be non-negative")
		os.Exit(1)
	}

	// Guard: Validate mode
	var stateMode state.Mode
	switch *mode {
	case "build":
		stateMode = state.ModeBuild
	case "plan":
		stateMode = state.ModePlan
	case "plan-work":
		// Guard: plan-work requires work description
		if *workDesc == "" {
			fmt.Fprintln(os.Stderr, "Error: plan-work mode requires --work flag")
			os.Exit(1)
		}
		stateMode = state.ModePlanWork
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid mode '%s'. Must be: build, plan, or plan-work\n", *mode)
		os.Exit(1)
	}

	// Initialize state and process manager
	appState := state.NewState()
	appState.SetMode(stateMode)
	appState.SetMaxIterations(*maxIter)
	appState.SetScriptPath(*scriptPath)
	if *workDesc != "" {
		appState.SetWorkDesc(*workDesc)
	}

	manager := process.NewManager(1000)

	// Create and run TUI
	model := tui.NewModel(appState, manager)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
