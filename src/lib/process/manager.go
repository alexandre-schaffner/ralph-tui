package process

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const (
	// DefaultBufferSize is the default ring buffer size for logs.
	DefaultBufferSize = 1000
)

// Status represents the current state of the managed process.
type Status int

const (
	StatusIdle Status = iota
	StatusRunning
	StatusStopping
	StatusStopped
	StatusPaused
)

func (s Status) String() string {
	switch s {
	case StatusIdle:
		return "Idle"
	case StatusRunning:
		return "Running"
	case StatusStopping:
		return "Stopping"
	case StatusStopped:
		return "Stopped"
	case StatusPaused:
		return "Paused"
	default:
		return "Unknown"
	}
}

// Manager manages a subprocess lifecycle with output streaming.
type Manager struct {
	cmd        *exec.Cmd
	status     Status
	logs       *RingBuffer
	doneChan   chan error
	mu         sync.RWMutex
	onComplete func() // Callback when process completes naturally
}

// NewManager creates a new process manager with a ring buffer for logs.
func NewManager(bufferSize int) *Manager {
	return &Manager{
		logs:     NewRingBuffer(bufferSize),
		doneChan: make(chan error, 1),
		status:   StatusIdle,
	}
}

// Start spawns the given command as a subprocess and begins streaming output.
// Returns error if process is already running or if command fails to start.
func (m *Manager) Start(command string, args ...string) error {
	m.mu.Lock()

	// Guard: Cannot start if already running or stopping
	if m.status == StatusRunning || m.status == StatusStopping {
		m.mu.Unlock()
		return fmt.Errorf("process already running or stopping")
	}

	// Close old doneChan if it exists to prevent leaks
	if m.doneChan != nil {
		select {
		case <-m.doneChan:
			// Already drained
		default:
			// Not drained, but we're starting fresh
		}
	}

	// Parse command into trusted state
	m.cmd = exec.Command(command, args...)
	m.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create process group for clean child termination
	}
	m.status = StatusRunning
	m.doneChan = make(chan error, 1)

	// Capture stdout and stderr
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		m.status = StatusIdle
		m.mu.Unlock()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := m.cmd.StderrPipe()
	if err != nil {
		m.status = StatusIdle
		m.mu.Unlock()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start process
	if err := m.cmd.Start(); err != nil {
		m.status = StatusIdle
		m.mu.Unlock()
		return fmt.Errorf("failed to start process: %w", err)
	}

	m.mu.Unlock()

	// Stream output in background goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go m.streamOutput(&wg, stdout, "[OUT]")
	go m.streamOutput(&wg, stderr, "[ERR]")

	// Wait for process completion in background
	go func() {
		wg.Wait() // Wait for output streams to close
		err := m.cmd.Wait()

		m.mu.Lock()
		m.status = StatusStopped
		// Copy callback under lock to prevent race
		callback := m.onComplete
		m.mu.Unlock()

		m.doneChan <- err

		// Trigger completion callback if registered
		if callback != nil {
			callback()
		}
	}()

	return nil
}

// streamOutput reads from the given reader and writes to the ring buffer.
func (m *Manager) streamOutput(wg *sync.WaitGroup, r io.Reader, prefix string) {
	defer wg.Done()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		m.logs.Write(fmt.Sprintf("%s %s", prefix, line))
	}
}

// Stop sends SIGTERM to the process and waits up to 5 seconds before sending SIGKILL.
func (m *Manager) Stop() error {
	m.mu.Lock()

	// Guard: Cannot stop if not running or already stopping
	if m.status != StatusRunning && m.status != StatusStopping {
		m.mu.Unlock()
		return fmt.Errorf("process not running")
	}

	if m.cmd == nil || m.cmd.Process == nil {
		m.mu.Unlock()
		return fmt.Errorf("no process to stop")
	}

	m.status = StatusStopping
	process := m.cmd.Process
	doneChan := m.doneChan
	m.mu.Unlock()

	// Send SIGTERM to process group for graceful shutdown
	pgid := process.Pid
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		// Handle ESRCH (no such process) gracefully
		if err == syscall.ESRCH {
			return nil
		}
		// Fallback to single process if process group kill fails
		if err := process.Signal(syscall.SIGTERM); err != nil {
			if err == syscall.ESRCH {
				return nil
			}
			return fmt.Errorf("failed to send SIGTERM: %w", err)
		}
	}

	// Wait for graceful shutdown with timeout
	select {
	case <-doneChan:
		return nil
	case <-time.After(5 * time.Second):
		// Timeout - force kill process group
		if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
			// Handle ESRCH gracefully
			if err != syscall.ESRCH {
				// Fallback to single process kill
				if err := process.Kill(); err != nil && err != syscall.ESRCH {
					return fmt.Errorf("failed to kill process: %w", err)
				}
			}
		}
		return fmt.Errorf("process killed after timeout")
	}
}

// StopImmediate sends SIGINT to the process for immediate interruption (<2s).
func (m *Manager) StopImmediate() error {
	m.mu.Lock()

	// Guard: Cannot stop if not running or already stopping
	if m.status != StatusRunning && m.status != StatusStopping {
		m.mu.Unlock()
		return fmt.Errorf("process not running")
	}

	if m.cmd == nil || m.cmd.Process == nil {
		m.mu.Unlock()
		return fmt.Errorf("no process to stop")
	}

	m.status = StatusStopping
	process := m.cmd.Process
	doneChan := m.doneChan
	m.mu.Unlock()

	// Send SIGINT to process group for immediate stop
	pgid := process.Pid
	if err := syscall.Kill(-pgid, syscall.SIGINT); err != nil {
		// Handle ESRCH (no such process) gracefully
		if err == syscall.ESRCH {
			return nil
		}
		// Fallback to single process if process group kill fails
		if err := process.Signal(syscall.SIGINT); err != nil {
			if err == syscall.ESRCH {
				return nil
			}
			return fmt.Errorf("failed to send SIGINT: %w", err)
		}
	}

	// Wait for immediate stop with 2 second timeout
	select {
	case <-doneChan:
		return nil
	case <-time.After(2 * time.Second):
		// Timeout - force kill process group
		if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
			// Handle ESRCH gracefully
			if err != syscall.ESRCH {
				// Fallback to single process kill
				if err := process.Kill(); err != nil && err != syscall.ESRCH {
					return fmt.Errorf("failed to kill process: %w", err)
				}
			}
		}
		return fmt.Errorf("process killed after timeout")
	}
}

// Pause stops the process gracefully and transitions to paused state.
// This allows manual restart later.
func (m *Manager) Pause() error {
	m.mu.Lock()

	// Guard: Cannot pause if not running
	if m.status != StatusRunning {
		m.mu.Unlock()
		return fmt.Errorf("process not running")
	}

	if m.cmd == nil || m.cmd.Process == nil {
		m.mu.Unlock()
		return fmt.Errorf("no process to pause")
	}

	m.status = StatusStopping
	process := m.cmd.Process
	doneChan := m.doneChan
	m.mu.Unlock()

	// Send SIGTERM to process group for graceful shutdown
	pgid := process.Pid
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		// Handle ESRCH (no such process) gracefully
		if err == syscall.ESRCH {
			m.mu.Lock()
			m.status = StatusPaused
			m.mu.Unlock()
			return nil
		}
		// Fallback to single process if process group kill fails
		if err := process.Signal(syscall.SIGTERM); err != nil {
			if err == syscall.ESRCH {
				m.mu.Lock()
				m.status = StatusPaused
				m.mu.Unlock()
				return nil
			}
			return fmt.Errorf("failed to send SIGTERM: %w", err)
		}
	}

	// Wait for graceful shutdown with timeout
	select {
	case <-doneChan:
		m.mu.Lock()
		m.status = StatusPaused
		m.mu.Unlock()
		return nil
	case <-time.After(5 * time.Second):
		// Timeout - force kill process group
		if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
			// Handle ESRCH gracefully
			if err != syscall.ESRCH {
				// Fallback to single process kill
				if err := process.Kill(); err != nil && err != syscall.ESRCH {
					return fmt.Errorf("failed to kill process: %w", err)
				}
			}
		}
		m.mu.Lock()
		m.status = StatusPaused
		m.mu.Unlock()
		return fmt.Errorf("process killed after timeout")
	}
}

// IsRunning returns true if the process is currently running.
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status == StatusRunning
}

// IsPaused returns true if the process is paused.
func (m *Manager) IsPaused() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status == StatusPaused
}

// GetStatus returns the current status of the process.
func (m *Manager) GetStatus() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// GetLogs returns all current log lines.
func (m *Manager) GetLogs() []string {
	return m.logs.ReadAll()
}

// ClearLogs empties the log buffer.
func (m *Manager) ClearLogs() {
	m.logs.Clear()
}

// OnComplete registers a callback to invoke when the process completes.
func (m *Manager) OnComplete(fn func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onComplete = fn
}

// WaitForExit blocks until the process exits and returns the error (if any).
func (m *Manager) WaitForExit() error {
	return <-m.doneChan
}
