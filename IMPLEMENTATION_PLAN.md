---
status: phase-3-complete
phase: 3
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
| Pause as Stop+Resume | Cannot truly pause bash script; graceful stop allows review between iterations. | `specs/prd-ralph-tui.md` |
| SIGINT for Immediate | SIGINT provides <2s interruption vs SIGTERM's graceful 5s timeout. | `specs/prd-ralph-tui.md` US-004 |

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

## ✓ Completed: Advanced Features & Polish (Phase 3)
- [x] **3.1 Immediate Stop (US-004)**
  - Added `StopImmediate()`: Sends SIGINT to process group for <2s interruption
  - Added 'X' keybinding for immediate stop vs 'x' graceful stop
  - ESRCH handling: Gracefully handle "no such process" errors
  - Guard in Start(): Prevent starting during StatusStopping
- [x] **3.2 Pause/Resume (US-005)**
  - Added `Pause()`: Graceful stop that transitions to StatusPaused
  - Added 'p' keybinding to pause running process
  - Resume via 's' key preserves iteration count and logs
  - Footer shows "s:resume" when paused
- [x] **3.3 Quit Confirmation (US-010)**
  - Show confirmation prompt when quitting with running process
  - Footer displays "Process is running. Quit anyway? (y/n)"
  - 'y' confirms and stops process before exit
  - 'n' cancels and returns to TUI
- [x] **3.4 Specs Browser Selection (US-008)**
  - Selectable file list with ↑↓ navigation and enter to view
  - Full-screen file viewer with scrolling support
  - Esc/backspace to return to file list
  - Highlight selected file in list
- [x] **3.5 Plan/Specs Scrolling (US-007, US-008)**
  - Added scroll offsets for plan and specs views
  - ↑↓ keys scroll line-by-line
  - PgUp/PgDn scroll 10 lines at a time
  - Header shows scroll hints
- [x] **3.6 Process Safety & Edge Cases**
  - Guard Start() against StatusStopping state
  - Handle ESRCH errors in Stop/StopImmediate/Pause
  - Safe process group termination with fallback
  - Detached HEAD handling: Show "detached@<hash>" instead of empty
- [x] **3.7 UI Enhancements**
  - Color-coded status indicators (green=running, yellow=stopping, blue=paused, red=stopped)
  - Dynamic footer showing context-appropriate keybindings
  - Pause notification uses info color vs error red
  - Dashboard shows iteration count for paused processes

## Future Enhancements (Phase 4 - Not Planned)
- [ ] Log filtering/search within TUI
- [ ] Configuration file support (~/.ralph-tui.toml)
- [ ] Multiple script profiles (save mode/max iterations presets)
- [ ] Export logs to file from TUI
- [ ] Real-time syntax highlighting for logs
