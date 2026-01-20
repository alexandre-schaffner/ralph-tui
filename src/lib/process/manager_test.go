package process

import (
	"strings"
	"testing"
	"time"
)

func TestManager_StartStop_Lifecycle(t *testing.T) {
	mgr := NewManager(DefaultBufferSize)

	// Guard: Should be idle initially
	if status := mgr.GetStatus(); status != StatusIdle {
		t.Errorf("Expected StatusIdle, got %v", status)
	}

	// Start a simple sleep command
	err := mgr.Start("sleep", "1")
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Should be running
	if !mgr.IsRunning() {
		t.Error("Expected process to be running")
	}

	// Stop the process
	err = mgr.Stop()
	if err != nil {
		t.Fatalf("Failed to stop process: %v", err)
	}

	// Should be stopped
	if mgr.GetStatus() != StatusStopped {
		t.Errorf("Expected StatusStopped, got %v", mgr.GetStatus())
	}
}

func TestManager_StopNonRunningProcess(t *testing.T) {
	mgr := NewManager(DefaultBufferSize)

	// Attempt to stop when not running
	err := mgr.Stop()
	if err == nil {
		t.Error("Expected error when stopping non-running process")
	}

	expectedMsg := "process not running"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestManager_PreventMultipleStarts(t *testing.T) {
	mgr := NewManager(DefaultBufferSize)

	// Start first process
	err := mgr.Start("sleep", "2")
	if err != nil {
		t.Fatalf("Failed to start first process: %v", err)
	}

	// Attempt to start second process
	err = mgr.Start("sleep", "1")
	if err == nil {
		t.Error("Expected error when starting process twice")
	}

	expectedMsg := "already running"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedMsg, err.Error())
	}

	// Clean up
	_ = mgr.Stop()
}

func TestManager_ProcessOutputCapture(t *testing.T) {
	mgr := NewManager(DefaultBufferSize)

	// Start echo command
	err := mgr.Start("sh", "-c", "echo 'test output'")
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Wait a bit for output to be captured
	time.Sleep(100 * time.Millisecond)

	// Check logs contain output
	logs := mgr.GetLogs()
	found := false
	for _, line := range logs {
		if strings.Contains(line, "test output") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected logs to contain 'test output'")
	}

	// Wait for process to complete
	_ = mgr.WaitForExit()
}

func TestManager_OnCompleteCallback(t *testing.T) {
	mgr := NewManager(DefaultBufferSize)

	callbackFired := false
	mgr.OnComplete(func() {
		callbackFired = true
	})

	// Start short-lived process
	err := mgr.Start("sh", "-c", "exit 0")
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Wait for completion
	err = mgr.WaitForExit()
	if err != nil {
		t.Errorf("Process exited with error: %v", err)
	}

	// Give callback time to fire
	time.Sleep(50 * time.Millisecond)

	if !callbackFired {
		t.Error("Expected OnComplete callback to fire")
	}
}

func TestManager_ClearLogs(t *testing.T) {
	mgr := NewManager(DefaultBufferSize)

	// Start process that generates output
	err := mgr.Start("sh", "-c", "echo 'line1'; echo 'line2'")
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Wait for output
	time.Sleep(100 * time.Millisecond)

	// Verify logs exist
	logs := mgr.GetLogs()
	if len(logs) == 0 {
		t.Error("Expected logs to be present")
	}

	// Clear logs
	mgr.ClearLogs()

	// Verify logs are cleared
	logs = mgr.GetLogs()
	if len(logs) != 0 {
		t.Errorf("Expected logs to be cleared, got %d lines", len(logs))
	}

	// Wait for process to complete
	_ = mgr.WaitForExit()
}

func TestManager_ProcessGroupTermination(t *testing.T) {
	mgr := NewManager(DefaultBufferSize)

	// Start a parent shell that spawns a child sleep process
	// This tests that the entire process group is terminated
	err := mgr.Start("sh", "-c", "sleep 30 & wait")
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	// Give it time to start child process
	time.Sleep(100 * time.Millisecond)

	// Stop should terminate both parent and child
	err = mgr.Stop()
	if err != nil && !strings.Contains(err.Error(), "killed after timeout") {
		t.Fatalf("Failed to stop process: %v", err)
	}

	// Verify process is stopped
	if mgr.IsRunning() {
		t.Error("Expected process to be stopped")
	}
}

func TestManager_RestartAfterStop(t *testing.T) {
	mgr := NewManager(DefaultBufferSize)

	// Start first process
	err := mgr.Start("sleep", "1")
	if err != nil {
		t.Fatalf("Failed to start first process: %v", err)
	}

	// Stop it
	err = mgr.Stop()
	if err != nil {
		t.Fatalf("Failed to stop process: %v", err)
	}

	// Wait a bit for cleanup
	time.Sleep(100 * time.Millisecond)

	// Start second process
	err = mgr.Start("sleep", "1")
	if err != nil {
		t.Fatalf("Failed to start second process: %v", err)
	}

	// Should be running
	if !mgr.IsRunning() {
		t.Error("Expected second process to be running")
	}

	// Clean up
	_ = mgr.Stop()
}
