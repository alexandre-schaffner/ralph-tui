---
status: phase-2-hardened
phase: 2
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

## ✓ Completed: Core Infrastructure & Safety (Phase 1)
- [x] **1.1 Initialize Project Structure**
  - Created `cmd/ralph-tui/main.go` (CLI entry)
  - Created `src/lib/process/` (Manager package)
  - Created `src/lib/state/` (State package)
  - Created `src/tui/` (UI package)
  - Initialized Go module with Bubbletea/Lipgloss
- [x] **1.2 Implement Process Manager (`src/lib/process`)**
  - Implemented `Start()`: Spawns `loop.sh` as subprocess with arguments
  - Implemented `Stop()`: Sends SIGTERM, escalates to SIGKILL after 5s timeout
  - Implemented streaming via goroutines reading stdout/stderr
  - Handles zombie processes via `Wait()` in background goroutine
- [x] **1.3 Implement Ring Buffer for Streaming**
  - Created fixed-size circular buffer (1000 lines) in `src/lib/process/ringbuffer.go`
  - Thread-safe via RWMutex for concurrent writes/reads
- [x] **1.4 Implement Centralized State Store (`src/lib/state`)**
  - Defined `State` struct with all required fields
  - Implemented thread-safe getters/setters using RWMutex
- [x] **1.5 Basic Bubbletea Scaffold (`src/tui`)**
  - Wired `main.go` to start Bubbletea program with flags
  - Implemented root model in `src/tui/model.go`
  - Connected State Store to Model Update loop
  - Implemented all 4 views: Dashboard, Logs, Plan, Specs
  - Added keyboard controls: s=start, x=stop, 1-4=tabs, q=quit
  - Terminal size check (min 80x24)

## ✓ Completed: UI Views & Features (Phase 2)
- [x] **2.1 Implement Logs View (`src/tui/views/logs`)**
  - Render content from Ring Buffer
  - Auto-scroll to bottom logic
- [x] **2.2 Connect Controls & Feedback**
  - Bind 's' key to Start, 'x' to Stop, 'q' to Quit
  - Update State.Status based on Process Manager events
  - Render "Running", "Stopped", "Stopping" indicators
- [x] **2.3 Implement Dashboard View (`src/tui/views/dashboard`)**
  - Parse streaming output for "LOOP X" patterns
  - Update and display Iteration Count in State
- [x] **2.4 Implement Plan & Specs Views**
  - `src/tui/views/plan`: Read/Render `IMPLEMENTATION_PLAN.md` with caching
  - `src/tui/views/specs`: Read/Render spec files with caching
  - Add Tab navigation (Dashboard | Logs | Plan | Specs)

## ✓ Completed: Hardening & Critical Fixes (Phase 2.5)
- [x] **2.5 Process Manager Hardening**
  - Fixed deadlock in Stop(): Release lock before waiting on doneChan
  - Fixed onComplete callback race: Copy under lock before calling
  - Implemented process group handling: Set Setpgid and kill entire process group
  - Fixed parseLogsForIterations never called: Now batched with tickForLogs
  - Fixed channel reinit without cleanup: Close old channels properly
  - Added comprehensive manager tests: lifecycle, stop non-running, prevent multiple starts, restart after stop
- [x] **2.6 Configuration & Usability**
  - Made loop.sh path configurable via --script flag (default: ./loop.sh)
  - Added validation for --max (must be non-negative)
  - Implemented file I/O caching for plan/specs views (5s cache duration)
  - Removed unused stopChan field
  - Extracted buffer size constant (DefaultBufferSize = 1000)

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
