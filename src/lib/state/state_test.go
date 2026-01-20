package state

import (
	"testing"

	"github.com/alex/ralph-tui/src/lib/process"
)

func TestState_ProcessStatus(t *testing.T) {
	s := NewState()

	if s.GetProcessStatus() != process.StatusIdle {
		t.Errorf("expected initial status Idle, got %v", s.GetProcessStatus())
	}

	s.SetProcessStatus(process.StatusRunning)
	if s.GetProcessStatus() != process.StatusRunning {
		t.Errorf("expected status Running, got %v", s.GetProcessStatus())
	}
}

func TestState_Mode(t *testing.T) {
	s := NewState()

	if s.GetMode() != ModeBuild {
		t.Errorf("expected default mode Build, got %v", s.GetMode())
	}

	s.SetMode(ModePlan)
	if s.GetMode() != ModePlan {
		t.Errorf("expected mode Plan, got %v", s.GetMode())
	}
}

func TestState_Iteration(t *testing.T) {
	s := NewState()

	if s.GetCurrentIteration() != 0 {
		t.Errorf("expected initial iteration 0, got %d", s.GetCurrentIteration())
	}

	s.IncrementIteration()
	s.IncrementIteration()

	if s.GetCurrentIteration() != 2 {
		t.Errorf("expected iteration 2, got %d", s.GetCurrentIteration())
	}

	s.ResetIteration()
	if s.GetCurrentIteration() != 0 {
		t.Errorf("expected iteration 0 after reset, got %d", s.GetCurrentIteration())
	}
}

func TestState_Error(t *testing.T) {
	s := NewState()

	if s.GetError() != "" {
		t.Errorf("expected no error initially, got %s", s.GetError())
	}

	s.SetError("test error")
	if s.GetError() != "test error" {
		t.Errorf("expected 'test error', got %s", s.GetError())
	}

	s.ClearError()
	if s.GetError() != "" {
		t.Errorf("expected no error after clear, got %s", s.GetError())
	}
}

func TestState_Complete(t *testing.T) {
	s := NewState()

	if s.GetComplete() {
		t.Error("expected incomplete initially")
	}

	s.SetComplete(true)
	if !s.GetComplete() {
		t.Error("expected complete after set")
	}
}
