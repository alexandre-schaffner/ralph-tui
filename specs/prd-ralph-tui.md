# PRD: Ralph TUI

## Introduction

Ralph TUI is a Terminal User Interface control center for managing the Ralph Wiggum autonomous AI coding agent. Currently, Ralph executes via `loop.sh`, which runs stateless iterations (read specs → implement → test → commit → repeat). Operators have zero visibility into real-time agent activity and no control beyond starting the script or killing the process.

This TUI provides operators with complete visibility and control over Ralph loop execution. Operators can monitor agent progress, manage iteration flow, and review planning artifacts without leaving the terminal.

## Goals

- Provide real-time visibility into AI agent activity during loop execution
- Enable operators to start, stop, and pause Ralph loops on demand
- Display current loop state (idle/planning/building) and iteration count
- Centralize access to specs and implementation plans
- Wrap existing `loop.sh` without reimplementing core logic

## User Stories

### US-001: View Real-Time Agent Output

**Description:** As a Ralph operator, I want to see live streaming output from the AI agent so that I understand what it's currently doing without checking log files.

**Acceptance Criteria:**
- [ ] Log view displays stdout/stderr from `loop.sh` subprocess in real-time
- [ ] New log entries appear automatically without manual refresh
- [ ] Log view scrolls to show historical output from current session
- [ ] Iteration boundaries display clearly (e.g., "LOOP 1", "LOOP 2")
- [ ] Output appears within 500ms of subprocess emission
- [ ] Typecheck passes

### US-002: Start a Ralph Loop

**Description:** As a Ralph operator, I want to start a Ralph loop from the TUI so that I don't need to exit to the command line.

**Acceptance Criteria:**
- [ ] "Start" action available when no loop is running
- [ ] User selects mode: build, plan, or plan-work
- [ ] User optionally sets max iterations (0 = unlimited)
- [ ] Loop starts as subprocess with selected parameters
- [ ] UI transitions to "running" state showing live output
- [ ] Start action is hidden when loop is already running
- [ ] Typecheck passes

### US-003: Stop a Running Loop (Graceful)

**Description:** As a Ralph operator, I want to stop a running loop gracefully so that the current iteration completes before shutdown.

**Acceptance Criteria:**
- [ ] "Stop (Graceful)" action available when loop is running
- [ ] Current iteration completes before subprocess terminates
- [ ] UI displays "Stopping..." status while waiting
- [ ] UI transitions to idle state once subprocess exits
- [ ] Final logs from iteration remain visible
- [ ] Typecheck passes

### US-004: Stop a Running Loop (Immediate)

**Description:** As a Ralph operator, I want to immediately interrupt a running loop so that I can stop execution when an iteration is taking too long or heading in the wrong direction.

**Acceptance Criteria:**
- [ ] "Stop (Immediate)" action available when loop is running
- [ ] Action sends SIGINT to subprocess immediately
- [ ] Subprocess terminates within 2 seconds
- [ ] UI transitions to idle state after termination
- [ ] UI displays confirmation message indicating interruption
- [ ] Typecheck passes

### US-005: Pause a Running Loop

**Description:** As a Ralph operator, I want to pause loop execution between iterations so that I can review progress before continuing.

**Acceptance Criteria:**
- [ ] "Pause" action available when loop is running
- [ ] Current iteration completes before pausing
- [ ] UI displays "Paused" status
- [ ] "Resume" action becomes available when paused
- [ ] Resume continues from next iteration
- [ ] Typecheck passes

### US-006: View Loop Status Dashboard

**Description:** As a Ralph operator, I want to see current loop state at a glance so that I always know what's happening.

**Acceptance Criteria:**
- [ ] Dashboard displays state: Idle, Running, Paused, or Stopping
- [ ] Dashboard shows mode when running (build/plan/plan-work)
- [ ] Dashboard displays current iteration number when running
- [ ] Dashboard shows max iterations if set (or "unlimited")
- [ ] Dashboard shows current git branch
- [ ] All fields update in real-time without manual refresh
- [ ] Typecheck passes

### US-007: View Implementation Plan

**Description:** As a Ralph operator, I want to read the current implementation plan so that I understand what the agent is working towards.

**Acceptance Criteria:**
- [ ] Plan viewer displays contents of `IMPLEMENTATION_PLAN.md` if it exists
- [ ] Plan viewer shows "No plan found" message if file doesn't exist
- [ ] Plan content renders as readable text (preserve formatting)
- [ ] Plan viewer scrolls for long plans
- [ ] Plan content refreshes when switching to plan view
- [ ] Typecheck passes

### US-008: Browse Spec Files

**Description:** As a Ralph operator, I want to browse and view specification files in `specs/*.md` so that I can reference requirements without leaving the TUI.

**Acceptance Criteria:**
- [ ] Spec browser lists all `.md` files in `specs/` directory
- [ ] Browser shows "No specs found" if directory is empty or doesn't exist
- [ ] User can select a spec file to view its contents
- [ ] Selected spec displays in scrollable viewer
- [ ] User can return to spec list from viewer
- [ ] Typecheck passes

### US-009: Navigate Between TUI Views

**Description:** As a Ralph operator, I want to switch between different views (logs, dashboard, plan, specs) so that I can access the information I need.

**Acceptance Criteria:**
- [ ] Tab or key navigation switches between views
- [ ] Available views: Dashboard, Logs, Plan, Specs
- [ ] Current view is clearly indicated (highlighted tab or label)
- [ ] Navigation works whether loop is running or idle
- [ ] Help text shows available navigation keys
- [ ] Typecheck passes

### US-010: Exit the TUI

**Description:** As a Ralph operator, I want to safely exit the TUI so that I can return to my shell without orphaning processes.

**Acceptance Criteria:**
- [ ] Quit action available at all times (e.g., 'q' key or Ctrl-C)
- [ ] If loop is running, user is prompted to confirm exit
- [ ] On confirmed exit, subprocess receives SIGTERM
- [ ] TUI waits up to 5 seconds for graceful subprocess shutdown
- [ ] TUI force-kills subprocess if timeout exceeded
- [ ] TUI exits cleanly to shell with zero exit code
- [ ] Typecheck passes

## Functional Requirements

**FR-1**: Spawn `loop.sh` as subprocess and capture stdout/stderr streams in real-time.

**FR-2**: Pass parameters to `loop.sh` matching its CLI: mode (build/plan/plan-work), max iterations, and work description for plan-work mode.

**FR-3**: Display subprocess output with <500ms latency from emission to display.

**FR-4**: Track subprocess state: not started, running, paused, stopping, stopped.

**FR-5**: Support graceful stop by allowing current iteration to complete before terminating subprocess.

**FR-6**: Support immediate stop by sending SIGINT to subprocess.

**FR-7**: Support pause by preventing next iteration from starting while allowing current one to complete.

**FR-8**: Display status dashboard showing: state, mode, iteration count, max iterations, git branch.

**FR-9**: Provide view for displaying contents of `IMPLEMENTATION_PLAN.md`.

**FR-10**: Provide file browser for `specs/*.md` files with ability to view selected files.

**FR-11**: Render correctly at minimum 80x24 terminal dimensions without layout breakage.

**FR-12**: Handle subprocess crashes by transitioning to stopped state and displaying error message.

**FR-13**: Detect loop completion by parsing `<promise>COMPLETE</promise>` from subprocess output and display completion status.

**FR-14**: Provide keyboard navigation between views: Dashboard, Logs, Plan, Specs.

**FR-15**: Clean up subprocess on exit: send SIGTERM, wait 5 seconds, send SIGKILL if necessary.

## Non-Goals

The following are explicitly **NOT** in scope:

- Git operations (viewing commits, branch management, history)
- Spec editing (view only)
- Plan editing (view only; edit externally in IDE)
- Configuration UI or settings screen
- History/analytics tracking from past loops or sessions
- Multi-project support (TUI runs in single project directory)
- Custom prompts (no selecting or editing `PROMPT_*.md` files)
- Model selection (respects loop.sh MODEL variable)
- Log persistence to disk (in-memory only for current session)
- Log filtering or searching
- Split screen or simultaneous multi-view display

## Technical Considerations

### Technology Stack
- **Language**: Go
- **TUI Framework**: Bubbletea (Charm libraries ecosystem)
- **Components**: Bubbles (Bubbletea components), Lipgloss (styling)

### Integration Points
- **loop.sh**: Wrapped as subprocess using Go's `os/exec` package
- **File System**: Read access to `IMPLEMENTATION_PLAN.md` and `specs/*.md`
- **Git**: Read git branch via `git branch --show-current` (subprocess)

### Subprocess Management
- Use `exec.Command` to spawn `loop.sh` with arguments
- Capture stdout/stderr via `StdoutPipe()` and `StderrPipe()`
- Use `cmd.Process.Signal()` for graceful termination (SIGTERM)
- Use `cmd.Process.Kill()` for immediate termination (SIGKILL)
- Handle cleanup in defer blocks to prevent orphans

### Pause Implementation
The TUI cannot pause `loop.sh` mid-execution (it's a bash script). For MVP, implement pause as graceful stop + manual restart. This provides the user benefit (review between iterations) without modifying `loop.sh`.

### Terminal Compatibility
- Assume standard terminal with ANSI support
- Minimum dimensions: 80x24
- Use ASCII + basic box-drawing characters (no unicode/emoji dependencies)

### Error Handling
- **Subprocess crash**: Display error state, show last logs, allow restart
- **Missing files**: Show "not found" messages for plan/specs, don't crash
- **Terminal resize**: Bubbletea handles automatically

## Success Metrics

**Primary Metric**: Operator can answer "What is Ralph doing right now?" without leaving the TUI or checking external logs.

**Validation Criteria**:
- Loop state (idle/running/paused) is visible at all times
- Current iteration number displays when running
- Live logs show agent output in real-time
- Operator can stop/pause loop within 2 interactions

**User Acceptance**: Operators prefer using TUI over running `./loop.sh` directly for all Ralph operations.

## Open Questions

**Q1**: Should pause be included in MVP given the technical complexity of pausing bash scripts?

**Recommendation**: Include pause in MVP but implement as "pause = graceful stop + manual restart" rather than true SIGSTOP/SIGCONT. This provides user benefit without requiring loop.sh modifications.

**Q2**: How should the TUI handle terminal dimensions below 80x24?

**Recommendation**: Display warning message asking user to resize terminal, similar to standard TUI application behavior.

**Q3**: Should log output be colorized/syntax highlighted or displayed as plain text?

**Recommendation**: Display as plain text for MVP. Preserve existing ANSI color codes from `loop.sh` output (which includes `opencode` colors). Don't add additional highlighting.

**Q4**: Should the TUI auto-refresh plan view when `IMPLEMENTATION_PLAN.md` changes on disk?

**Recommendation**: No for MVP. Refresh only when user navigates to plan view. File watching adds complexity and plan typically changes between iterations (when user views logs).

**Q5**: What should happen if user tries to start a loop while one is already running?

**Recommendation**: Disable/hide start action when loop is running. Prevent scenario via UI state management.
