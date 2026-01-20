---
status: not-started
phase: 1
updated: 2026-01-20
---

# Implementation Plan

## Goal
Create a TUI that starts/stops `loop.sh` with <500ms response, streams stdout/stderr with <100ms lag, displays iteration count, and handles process crashes gracefully.

## Context & Decisions
| Decision | Rationale | Source |
|----------|-----------|--------|
| `loop.sh` wrapping | Reuse existing bash logic; avoid reimplementation. | `specs/prd-ralph-tui.md` |
| Bubbletea/Lipgloss | Standard Go TUI stack for robust event loops. | `specs/prd-ralph-tui.md` |
| Centralized State | Single source of truth (Redux-like) for UI consistency. | `AGENTS.md` |
| Ring Buffer | Prevent UI lag/crashes from massive log volume. | `reviewer-feedback` |

## High Priority: Core Infrastructure & Safety (Phase 1)
- [ ] **1.1 Initialize Project Structure**
  - Create `cmd/ralph-tui/main.go` (CLI entry)
  - Create `src/lib/process/` (Manager package)
  - Create `src/lib/state/` (State package)
  - Create `src/tui/` (UI package)
  - Run `go mod init`
- [ ] **1.2 Implement Process Manager (`src/lib/process`)**
  - Implement `Start(cmd string)`: Spawn `loop.sh` as subprocess
  - Implement `Stop()`: Send SIGTERM, escalate to SIGKILL after timeout
  - Implement `Stream()`: Return channels for stdout/stderr
  - **Constraint**: Handle zombie processes via `Wait()` in goroutine
- [ ] **1.3 Implement Ring Buffer for Streaming**
  - Create fixed-size circular buffer (e.g., 1000 lines) in `src/lib/process`
  - Ensure thread-safe writes from subprocess and reads from UI
- [ ] **1.4 Implement Centralized State Store (`src/lib/state`)**
  - Define `State` struct (Status, Logs slice, IterationCount)
  - Implement thread-safe `Update` methods (Mutex or Channels)
- [ ] **1.5 Basic Bubbletea Scaffold (`src/tui`)**
  - Wire `main.go` to start Bubbletea program
  - Implement root model in `src/tui/app.go`
  - Connect State Store to Model Update loop

## Medium Priority: UI Views & Features (Phase 2)
- [ ] **2.1 Implement Logs View (`src/tui/views/logs`)**
  - Render content from Ring Buffer
  - Auto-scroll to bottom logic
- [ ] **2.2 Connect Controls & Feedback**
  - Bind 's' key to Start, 'q' to Stop/Quit
  - Update State.Status based on Process Manager events
  - Render "Running", "Stopped", "Stopping" indicators
- [ ] **2.3 Implement Dashboard View (`src/tui/views/dashboard`)**
  - Parse streaming output for "Iteration: X" patterns
  - Update and display Iteration Count in State
- [ ] **2.4 Implement Plan & Specs Views**
  - `src/tui/views/plan`: Read/Render `IMPLEMENTATION_PLAN.md`
  - `src/tui/views/specs`: Read/Render `specs/prd-ralph-tui.md`
  - Add Tab navigation (Dashboard | Logs | Plan | Specs)

## Low Priority: Polish & Production (Phase 3)
- [ ] **3.1 Configuration & CLI**
  - Parse flags: `--loop-script`, `--log-level`
  - Support `RALPH_TUI_PORT` env var (if needed for remote)
- [ ] **3.2 Visual Styling (Lipgloss)**
  - Apply borders, padding, and colors to Views
  - Style status indicators (Green=Running, Red=Stopped)
- [ ] **3.3 Graceful Shutdown**
  - Trap SIGINT/SIGTERM in `main.go`
  - Ensure child process is killed before TUI exits
