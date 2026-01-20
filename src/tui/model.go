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

// fileCache caches file content with timestamp for periodic refresh.
type fileCache struct {
	content   string
	timestamp time.Time
}

// Model is the root Bubbletea model.
type Model struct {
	state             *state.State
	manager           *process.Manager
	width             int
	height            int
	ready             bool
	planCache         *fileCache
	specsCache        map[string]*fileCache
	cacheDuration     time.Duration
	showQuitConfirm   bool
	selectedSpecIndex int
	specsListCache    []string
	specsViewingFile  bool
	planScrollOffset  int
	specsScrollOffset int
}

// NewModel creates a new TUI model.
func NewModel(st *state.State, mgr *process.Manager) *Model {
	return &Model{
		state:             st,
		manager:           mgr,
		specsCache:        make(map[string]*fileCache),
		cacheDuration:     5 * time.Second, // Refresh cache every 5 seconds
		showQuitConfirm:   false,
		selectedSpecIndex: 0,
		specsViewingFile:  false,
		planScrollOffset:  0,
		specsScrollOffset: 0,
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchGitBranch(),
		m.tickForLogs(),
	)
}

// Update handles messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tickMsg:
		// Parse logs for iteration markers and batch with next tick
		parseCmd := m.parseLogsForIterations()
		tickCmd := m.tickForLogs()
		if parseCmd != nil {
			return m, tea.Batch(parseCmd, tickCmd)
		}
		return m, tickCmd

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
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle quit confirmation dialog
	if m.showQuitConfirm {
		switch msg.String() {
		case "y", "Y":
			return m, m.confirmQuit()
		case "n", "N", "q", "ctrl+c":
			m.showQuitConfirm = false
			return m, nil
		}
		return m, nil
	}

	// Handle specs view navigation
	if m.state.GetCurrentView() == "specs" {
		if m.specsViewingFile {
			switch msg.String() {
			case "esc", "backspace":
				m.specsViewingFile = false
				m.specsScrollOffset = 0
				return m, nil
			case "up", "k":
				if m.specsScrollOffset > 0 {
					m.specsScrollOffset--
				}
				return m, nil
			case "down", "j":
				m.specsScrollOffset++
				return m, nil
			case "pgup":
				m.specsScrollOffset -= 10
				if m.specsScrollOffset < 0 {
					m.specsScrollOffset = 0
				}
				return m, nil
			case "pgdown":
				m.specsScrollOffset += 10
				return m, nil
			}
		} else {
			switch msg.String() {
			case "up", "k":
				if m.selectedSpecIndex > 0 {
					m.selectedSpecIndex--
				}
				return m, nil
			case "down", "j":
				if len(m.specsListCache) > 0 && m.selectedSpecIndex < len(m.specsListCache)-1 {
					m.selectedSpecIndex++
				}
				return m, nil
			case "enter":
				if len(m.specsListCache) > 0 {
					m.specsViewingFile = true
					m.specsScrollOffset = 0
				}
				return m, nil
			}
		}
	}

	// Handle plan view scrolling
	if m.state.GetCurrentView() == "plan" {
		switch msg.String() {
		case "up", "k":
			if m.planScrollOffset > 0 {
				m.planScrollOffset--
			}
			return m, nil
		case "down", "j":
			m.planScrollOffset++
			return m, nil
		case "pgup":
			m.planScrollOffset -= 10
			if m.planScrollOffset < 0 {
				m.planScrollOffset = 0
			}
			return m, nil
		case "pgdown":
			m.planScrollOffset += 10
			return m, nil
		}
	}

	// Global key handlers
	switch msg.String() {
	case "ctrl+c", "q":
		return m, m.handleQuit()

	case "s":
		return m, m.handleStart()

	case "x":
		return m, m.handleStop()

	case "X":
		return m, m.handleStopImmediate()

	case "p":
		return m, m.handlePause()

	case "1":
		m.state.SetCurrentView("dashboard")
		m.specsViewingFile = false
		return m, nil

	case "2":
		m.state.SetCurrentView("logs")
		m.specsViewingFile = false
		return m, nil

	case "3":
		m.state.SetCurrentView("plan")
		// Invalidate cache on view switch for fresh content
		m.planCache = nil
		m.planScrollOffset = 0
		m.specsViewingFile = false
		return m, nil

	case "4":
		m.state.SetCurrentView("specs")
		// Invalidate cache on view switch for fresh content
		m.specsCache = make(map[string]*fileCache)
		m.specsListCache = nil
		m.selectedSpecIndex = 0
		m.specsScrollOffset = 0
		m.specsViewingFile = false
		return m, nil
	}

	return m, nil
}

// handleStart starts or resumes the loop process.
func (m *Model) handleStart() tea.Cmd {
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

	// Only reset iteration if not resuming from pause
	if !m.manager.IsPaused() {
		m.state.ResetIteration()
		m.manager.ClearLogs()
	}

	m.state.ClearError()
	m.state.SetComplete(false)

	scriptPath := m.state.GetScriptPath()
	err := m.manager.Start(scriptPath, args...)
	if err != nil {
		m.state.SetError(err.Error())
	} else {
		m.state.SetProcessStatus(process.StatusRunning)
	}

	return nil
}

// handleStop stops the loop process.
func (m *Model) handleStop() tea.Cmd {
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

// handleQuit handles application exit with confirmation if process running.
func (m *Model) handleQuit() tea.Cmd {
	// If process is running, show confirmation
	if m.manager.IsRunning() {
		m.showQuitConfirm = true
		return nil
	}
	return tea.Quit
}

// confirmQuit performs the actual quit after confirmation.
func (m *Model) confirmQuit() tea.Cmd {
	if m.manager.IsRunning() {
		_ = m.manager.Stop()
	}
	return tea.Quit
}

// handleStopImmediate sends SIGINT to the process for immediate stop.
func (m *Model) handleStopImmediate() tea.Cmd {
	if !m.manager.IsRunning() {
		return nil
	}

	err := m.manager.StopImmediate()
	if err != nil {
		m.state.SetError(err.Error())
	} else {
		m.state.SetError("Process interrupted (SIGINT)")
	}
	m.state.SetProcessStatus(process.StatusStopped)

	return nil
}

// handlePause pauses the loop process.
func (m *Model) handlePause() tea.Cmd {
	if !m.manager.IsRunning() {
		return nil
	}

	err := m.manager.Pause()
	if err != nil {
		m.state.SetError(err.Error())
	} else {
		m.state.SetError("Process paused - press 's' to resume")
	}
	m.state.SetProcessStatus(process.StatusPaused)

	return nil
}

// View renders the UI.
func (m *Model) View() string {
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
func (m *Model) renderSizeWarning() string {
	warning := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9")).
		Render("Terminal too small!")

	msg := fmt.Sprintf("\n%s\n\nMinimum size: 80x24\nCurrent size: %dx%d\n\nPlease resize your terminal.\n",
		warning, m.width, m.height)

	return msg
}

// renderMain renders the main UI layout.
func (m *Model) renderMain() string {
	header := m.renderHeader()
	tabs := m.renderTabs()
	content := m.renderContent()
	footer := m.renderFooter()

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, tabs, content, footer)
}

// renderHeader renders the application header.
func (m *Model) renderHeader() string {
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
func (m *Model) renderTabs() string {
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
func (m *Model) renderContent() string {
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
func (m *Model) renderDashboard(height int) string {
	var lines []string

	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Status Dashboard"))
	lines = append(lines, "")

	// Process status with color
	status := m.manager.GetStatus()
	statusStr := status.String()
	statusStyle := lipgloss.NewStyle()

	switch status {
	case process.StatusRunning:
		statusStyle = statusStyle.Foreground(lipgloss.Color("10"))
	case process.StatusStopping:
		statusStyle = statusStyle.Foreground(lipgloss.Color("11"))
	case process.StatusPaused:
		statusStyle = statusStyle.Foreground(lipgloss.Color("12"))
	case process.StatusStopped:
		statusStyle = statusStyle.Foreground(lipgloss.Color("9"))
	}

	lines = append(lines, fmt.Sprintf("Process Status: %s", statusStyle.Render(statusStr)))

	// Mode
	mode := m.state.GetMode()
	lines = append(lines, fmt.Sprintf("Mode: %s", mode))

	// Iteration count
	if m.manager.IsRunning() || m.manager.IsPaused() {
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

	// Error message (used for pause notification and errors)
	if errMsg := m.state.GetError(); errMsg != "" {
		lines = append(lines, "")
		if strings.Contains(errMsg, "paused") {
			lines = append(lines, lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Render(errMsg))
		} else {
			lines = append(lines, lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Render(errMsg))
		}
	}

	return strings.Join(lines, "\n")
}

// renderLogs renders the logs view.
func (m *Model) renderLogs(height int) string {
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

// renderPlan renders the plan view with scrolling support.
func (m *Model) renderPlan(height int) string {
	// Check cache validity
	now := time.Now()
	var content string

	if m.planCache != nil && now.Sub(m.planCache.timestamp) < m.cacheDuration {
		// Use cached content
		content = m.planCache.content
	} else {
		// Cache miss or expired - read from disk
		data, err := os.ReadFile("IMPLEMENTATION_PLAN.md")
		if err != nil {
			return "No implementation plan found.\n\nRun './loop.sh plan' to create one."
		}

		content = string(data)

		// Update cache
		m.planCache = &fileCache{
			content:   content,
			timestamp: now,
		}
	}

	var headerLines []string
	headerLines = append(headerLines, lipgloss.NewStyle().Bold(true).Render("Implementation Plan"))
	headerLines = append(headerLines, lipgloss.NewStyle().Faint(true).Render("(↑↓:scroll, pgup/pgdn:fast scroll)"))
	headerLines = append(headerLines, "")

	lines := strings.Split(content, "\n")

	// Apply scroll offset
	start := m.planScrollOffset
	if start >= len(lines) {
		start = len(lines) - 1
	}
	if start < 0 {
		start = 0
	}

	remaining := height - len(headerLines)
	end := start + remaining
	if end > len(lines) {
		end = len(lines)
	}

	result := append(headerLines, lines[start:end]...)
	return strings.Join(result, "\n")
}

// renderSpecs renders the specs view with selectable file list.
func (m *Model) renderSpecs(height int) string {
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

	// Update cache
	m.specsListCache = mdFiles

	// If viewing a file, show its content
	if m.specsViewingFile && m.selectedSpecIndex < len(mdFiles) {
		selectedFile := mdFiles[m.selectedSpecIndex]
		now := time.Now()
		var content string

		// Check cache validity
		if cache, exists := m.specsCache[selectedFile]; exists && now.Sub(cache.timestamp) < m.cacheDuration {
			content = cache.content
		} else {
			// Cache miss or expired - read from disk
			data, err := os.ReadFile(fmt.Sprintf("specs/%s", selectedFile))
			if err == nil {
				content = string(data)
				// Update cache
				m.specsCache[selectedFile] = &fileCache{
					content:   content,
					timestamp: now,
				}
			}
		}

		var lines []string
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Viewing: %s", selectedFile)))
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render("(esc:back, ↑↓:scroll, pgup/pgdn:fast scroll)"))
		lines = append(lines, "")

		if content != "" {
			contentLines := strings.Split(content, "\n")
			// Apply scroll offset
			start := m.specsScrollOffset
			if start >= len(contentLines) {
				start = len(contentLines) - 1
			}
			if start < 0 {
				start = 0
			}

			remaining := height - 3
			end := start + remaining
			if end > len(contentLines) {
				end = len(contentLines)
			}

			lines = append(lines, contentLines[start:end]...)
		}

		return strings.Join(lines, "\n")
	}

	// Show file list with selection
	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Specification Files"))
	lines = append(lines, lipgloss.NewStyle().Faint(true).Render("(↑↓:select, enter:view)"))
	lines = append(lines, "")

	for i, file := range mdFiles {
		if i == m.selectedSpecIndex {
			lines = append(lines, lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Render(fmt.Sprintf("  > %s", file)))
		} else {
			lines = append(lines, fmt.Sprintf("    %s", file))
		}
	}

	return strings.Join(lines, "\n")
}

// renderFooter renders the footer with keybindings.
func (m *Model) renderFooter() string {
	// Show quit confirmation if active
	if m.showQuitConfirm {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("Process is running. Quit anyway? (y/n)")
	}

	var keys []string

	if m.manager.IsRunning() {
		keys = append(keys, "x:stop(graceful)", "X:stop(immediate)", "p:pause")
	} else if m.manager.IsPaused() {
		keys = append(keys, "s:resume")
	} else {
		keys = append(keys, "s:start")
	}

	keys = append(keys, "1-4:tabs", "q:quit")

	return lipgloss.NewStyle().
		Faint(true).
		Render(strings.Join(keys, " | "))
}

// fetchGitBranch fetches the current git branch.
func (m *Model) fetchGitBranch() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "branch", "--show-current")
		output, err := cmd.Output()
		if err != nil {
			return gitBranchMsg("unknown")
		}
		branch := strings.TrimSpace(string(output))
		// Handle detached HEAD state
		if branch == "" {
			// Try to get the commit hash
			cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
			output, err = cmd.Output()
			if err != nil {
				return gitBranchMsg("detached HEAD")
			}
			branch = "detached@" + strings.TrimSpace(string(output))
		}
		return gitBranchMsg(branch)
	}
}

// tickForLogs polls for new logs and parses iteration markers.
func (m *Model) tickForLogs() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// parseLogsForIterations checks logs for iteration markers and completion.
func (m *Model) parseLogsForIterations() tea.Cmd {
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
