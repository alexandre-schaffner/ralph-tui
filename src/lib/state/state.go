package state

import (
	"sync"

	"github.com/alex/ralph-tui/src/lib/process"
)

// Mode represents the loop execution mode.
type Mode string

const (
	ModeBuild    Mode = "build"
	ModePlan     Mode = "plan"
	ModePlanWork Mode = "plan-work"
)

// State represents the centralized application state.
type State struct {
	// Process state
	ProcessStatus process.Status
	Mode          Mode
	MaxIterations int
	WorkDesc      string // For plan-work mode
	ScriptPath    string // Path to loop.sh script

	// Runtime state
	CurrentIteration int
	GitBranch        string
	ErrorMessage     string
	IsComplete       bool

	// UI state
	CurrentView  string // "dashboard", "logs", "plan", "specs"
	SelectedSpec string // For specs browser

	mu sync.RWMutex
}

// NewState creates a new application state.
func NewState() *State {
	return &State{
		ProcessStatus: process.StatusIdle,
		Mode:          ModeBuild,
		MaxIterations: 0,
		ScriptPath:    "./loop.sh",
		CurrentView:   "dashboard",
	}
}

// SetProcessStatus updates the process status.
func (s *State) SetProcessStatus(status process.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ProcessStatus = status
}

// GetProcessStatus returns the current process status.
func (s *State) GetProcessStatus() process.Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ProcessStatus
}

// SetMode updates the execution mode.
func (s *State) SetMode(mode Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Mode = mode
}

// GetMode returns the current mode.
func (s *State) GetMode() Mode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Mode
}

// SetMaxIterations updates the max iterations.
func (s *State) SetMaxIterations(max int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MaxIterations = max
}

// GetMaxIterations returns the max iterations.
func (s *State) GetMaxIterations() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MaxIterations
}

// SetWorkDesc updates the work description for plan-work mode.
func (s *State) SetWorkDesc(desc string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.WorkDesc = desc
}

// GetWorkDesc returns the work description.
func (s *State) GetWorkDesc() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.WorkDesc
}

// SetScriptPath updates the script path.
func (s *State) SetScriptPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ScriptPath = path
}

// GetScriptPath returns the script path.
func (s *State) GetScriptPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ScriptPath
}

// IncrementIteration increments the current iteration count.
func (s *State) IncrementIteration() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentIteration++
}

// GetCurrentIteration returns the current iteration.
func (s *State) GetCurrentIteration() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentIteration
}

// ResetIteration resets the iteration counter.
func (s *State) ResetIteration() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentIteration = 0
}

// SetGitBranch updates the git branch.
func (s *State) SetGitBranch(branch string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GitBranch = branch
}

// GetGitBranch returns the current git branch.
func (s *State) GetGitBranch() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.GitBranch
}

// SetError updates the error message.
func (s *State) SetError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ErrorMessage = msg
}

// GetError returns the error message.
func (s *State) GetError() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ErrorMessage
}

// ClearError clears the error message.
func (s *State) ClearError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ErrorMessage = ""
}

// SetComplete marks the loop as complete.
func (s *State) SetComplete(complete bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsComplete = complete
}

// GetComplete returns whether the loop is complete.
func (s *State) GetComplete() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IsComplete
}

// SetCurrentView updates the active UI view.
func (s *State) SetCurrentView(view string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentView = view
}

// GetCurrentView returns the active UI view.
func (s *State) GetCurrentView() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentView
}

// SetSelectedSpec updates the selected spec file.
func (s *State) SetSelectedSpec(spec string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SelectedSpec = spec
}

// GetSelectedSpec returns the selected spec file.
func (s *State) GetSelectedSpec() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SelectedSpec
}
