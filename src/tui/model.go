package tui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/alex/ralph-tui/src/lib/process"
	"github.com/alex/ralph-tui/src/lib/state"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the root Bubbletea model.
type Model struct {
	state   *state.State
	manager *process.Manager
	width   int
	height  int
	ready   bool
}

// NewModel creates a new TUI model.
func NewModel(st *state.State, mgr *process.Manager) Model {
	return Model{
		state:   st,
		manager: mgr,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchGitBranch(),
		m.tickForLogs(),
	)
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tickMsg:
		return m, m.tickForLogs()

	case gitBranchMsg:
		m.state.SetGitBranch(string(msg))
		return m, nil

	case iterationMsg:
		m.state.IncrementIteration()
		return m, nil

	case completeMsg:
		m.state.SetComplete(true)
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, m.handleQuit()

	case "s":
		return m, m.handleStart()

	case "x":
		return m, m.handleStop()

	case "1":
		m.state.SetCurrentView("dashboard")
		return m, nil

	case "2":
		m.state.SetCurrentView("logs")
		return m, nil

	case "3":
		m.state.SetCurrentView("plan")
		return m, nil

	case "4":
		m.state.SetCurrentView("specs")
		return m, nil
	}

	return m, nil
}

// handleStart starts the loop process.
func (m Model) handleStart() tea.Cmd {
	if m.manager.IsRunning() {
		return nil
	}

	mode := m.state.GetMode()
	maxIter := m.state.GetMaxIterations()

	args := []string{}
	if mode == state.ModePlan {
		args = append(args, "plan")
		if maxIter > 0 {
			args = append(args, fmt.Sprintf("%d", maxIter))
		}
	} else if mode == state.ModePlanWork {
		args = append(args, "plan-work")
		args = append(args, m.state.GetWorkDesc())
		if maxIter > 0 {
			args = append(args, fmt.Sprintf("%d", maxIter))
		}
	} else {
		// Build mode
		if maxIter > 0 {
			args = append(args, fmt.Sprintf("%d", maxIter))
		}
	}

	m.state.ResetIteration()
	m.state.ClearError()
	m.state.SetComplete(false)
	m.manager.ClearLogs()

	err := m.manager.Start("./loop.sh", args...)
	if err != nil {
		m.state.SetError(err.Error())
	} else {
		m.state.SetProcessStatus(process.StatusRunning)
	}

	return nil
}

// handleStop stops the loop process.
func (m Model) handleStop() tea.Cmd {
	if !m.manager.IsRunning() {
		return nil
	}

	err := m.manager.Stop()
	if err != nil {
		m.state.SetError(err.Error())
	}
	m.state.SetProcessStatus(process.StatusStopped)

	return nil
}

// handleQuit handles application exit.
func (m Model) handleQuit() tea.Cmd {
	if m.manager.IsRunning() {
		_ = m.manager.Stop()
	}
	return tea.Quit
}

// View renders the UI.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Guard: Enforce minimum terminal size
	if m.width < 80 || m.height < 24 {
		return m.renderSizeWarning()
	}

	return m.renderMain()
}

// renderSizeWarning displays a warning for undersized terminals.
func (m Model) renderSizeWarning() string {
	warning := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9")).
		Render("Terminal too small!")

	msg := fmt.Sprintf("\n%s\n\nMinimum size: 80x24\nCurrent size: %dx%d\n\nPlease resize your terminal.\n",
		warning, m.width, m.height)

	return msg
}

// renderMain renders the main UI layout.
func (m Model) renderMain() string {
	header := m.renderHeader()
	tabs := m.renderTabs()
	content := m.renderContent()
	footer := m.renderFooter()

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, tabs, content, footer)
}

// renderHeader renders the application header.
func (m Model) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		Render("RALPH TUI")

	branch := m.state.GetGitBranch()
	if branch == "" {
		branch = "unknown"
	}

	status := m.manager.GetStatus().String()
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	if m.manager.GetStatus() == process.StatusRunning {
		statusStyle = statusStyle.Foreground(lipgloss.Color("10"))
	}

	info := fmt.Sprintf("Branch: %s | Status: %s", branch, statusStyle.Render(status))

	return fmt.Sprintf("%s    %s", title, info)
}

// renderTabs renders the tab navigation.
func (m Model) renderTabs() string {
	currentView := m.state.GetCurrentView()

	tabs := []string{"1:Dashboard", "2:Logs", "3:Plan", "4:Specs"}
	views := []string{"dashboard", "logs", "plan", "specs"}

	var rendered []string
	for i, tab := range tabs {
		style := lipgloss.NewStyle().Padding(0, 2)
		if views[i] == currentView {
			style = style.Bold(true).Foreground(lipgloss.Color("12"))
		}
		rendered = append(rendered, style.Render(tab))
	}

	return strings.Join(rendered, " ")
}

// renderContent renders the current view content.
func (m Model) renderContent() string {
	contentHeight := m.height - 6 // Reserve space for header, tabs, footer

	switch m.state.GetCurrentView() {
	case "dashboard":
		return m.renderDashboard(contentHeight)
	case "logs":
		return m.renderLogs(contentHeight)
	case "plan":
		return m.renderPlan(contentHeight)
	case "specs":
		return m.renderSpecs(contentHeight)
	default:
		return "Unknown view"
	}
}

// renderDashboard renders the dashboard view.
func (m Model) renderDashboard(height int) string {
	var lines []string

	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Status Dashboard"))
	lines = append(lines, "")

	// Process status
	status := m.manager.GetStatus().String()
	lines = append(lines, fmt.Sprintf("Process Status: %s", status))

	// Mode
	mode := m.state.GetMode()
	lines = append(lines, fmt.Sprintf("Mode: %s", mode))

	// Iteration count
	if m.manager.IsRunning() {
		iter := m.state.GetCurrentIteration()
		maxIter := m.state.GetMaxIterations()
		if maxIter > 0 {
			lines = append(lines, fmt.Sprintf("Iteration: %d/%d", iter, maxIter))
		} else {
			lines = append(lines, fmt.Sprintf("Iteration: %d (unlimited)", iter))
		}
	}

	// Work description for plan-work mode
	if mode == state.ModePlanWork {
		lines = append(lines, fmt.Sprintf("Work: %s", m.state.GetWorkDesc()))
	}

	// Git branch
	branch := m.state.GetGitBranch()
	lines = append(lines, fmt.Sprintf("Branch: %s", branch))

	// Completion status
	if m.state.GetComplete() {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Render("✓ Loop completed successfully!"))
	}

	// Error message
	if errMsg := m.state.GetError(); errMsg != "" {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(fmt.Sprintf("Error: %s", errMsg)))
	}

	return strings.Join(lines, "\n")
}

// renderLogs renders the logs view.
func (m Model) renderLogs(height int) string {
	logs := m.manager.GetLogs()

	if len(logs) == 0 {
		return "No logs yet. Press 's' to start the loop."
	}

	// Show last N lines that fit in view
	start := 0
	if len(logs) > height {
		start = len(logs) - height
	}

	return strings.Join(logs[start:], "\n")
}

// renderPlan renders the plan view.
func (m Model) renderPlan(height int) string {
	data, err := os.ReadFile("IMPLEMENTATION_PLAN.md")
	if err != nil {
		return "No implementation plan found.\n\nRun './loop.sh plan' to create one."
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Show first N lines that fit
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// renderSpecs renders the specs view.
func (m Model) renderSpecs(height int) string {
	entries, err := os.ReadDir("specs")
	if err != nil {
		return "No specs directory found."
	}

	var mdFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			mdFiles = append(mdFiles, entry.Name())
		}
	}

	if len(mdFiles) == 0 {
		return "No spec files found in specs/"
	}

	// For MVP, just list the files
	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Specification Files"))
	lines = append(lines, "")
	for _, file := range mdFiles {
		lines = append(lines, fmt.Sprintf("  - %s", file))
	}

	// Show first spec content if available
	if len(mdFiles) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Preview: %s", mdFiles[0])))
		lines = append(lines, "")

		data, err := os.ReadFile(fmt.Sprintf("specs/%s", mdFiles[0]))
		if err == nil {
			preview := strings.Split(string(data), "\n")
			remaining := height - len(lines) - 1
			if len(preview) > remaining {
				preview = preview[:remaining]
			}
			lines = append(lines, preview...)
		}
	}

	return strings.Join(lines, "\n")
}

// renderFooter renders the footer with keybindings.
func (m Model) renderFooter() string {
	var keys []string

	if m.manager.IsRunning() {
		keys = append(keys, "x:stop")
	} else {
		keys = append(keys, "s:start")
	}

	keys = append(keys, "1-4:tabs", "q:quit")

	return lipgloss.NewStyle().
		Faint(true).
		Render(strings.Join(keys, " | "))
}

// fetchGitBranch fetches the current git branch.
func (m Model) fetchGitBranch() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "branch", "--show-current")
		output, err := cmd.Output()
		if err != nil {
			return gitBranchMsg("unknown")
		}
		return gitBranchMsg(strings.TrimSpace(string(output)))
	}
}

// tickForLogs polls for new logs and parses iteration markers.
func (m Model) tickForLogs() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// parseLogsForIterations checks logs for iteration markers and completion.
func (m Model) parseLogsForIterations() tea.Cmd {
	logs := m.manager.GetLogs()
	if len(logs) == 0 {
		return nil
	}

	// Check last few lines for iteration markers
	iterRegex := regexp.MustCompile(`LOOP (\d+)`)
	completeRegex := regexp.MustCompile(`<promise>COMPLETE</promise>`)

	start := len(logs) - 10
	if start < 0 {
		start = 0
	}

	for _, line := range logs[start:] {
		if matches := iterRegex.FindStringSubmatch(line); matches != nil {
			return func() tea.Msg { return iterationMsg{} }
		}
		if completeRegex.MatchString(line) {
			return func() tea.Msg { return completeMsg{} }
		}
	}

	return nil
}

// Message types
type tickMsg time.Time
type gitBranchMsg string
type iterationMsg struct{}
type completeMsg struct{}
