# Architectural Context

## Project Pattern
- **Wrapper TUI**: This application is a terminal user interface wrapper around an existing bash script (`loop.sh`).
- **State Management**: Uses a centralized state store (`src/lib/state`) with thread-safe access passed to Bubbletea models.
- **Concurrency**: `loop.sh` runs in a subprocess; stdout/stderr are streamed to a ring buffer via goroutines.

## Key Gotchas
- **Process Control**: Uses process groups (Setpgid=true) to terminate `loop.sh` and all child processes (SIGTERM to -pgid → 5s → SIGKILL to -pgid).
- **Streaming**: Large output from `loop.sh` uses a 1000-line ring buffer to prevent UI lag.
- **Thread Safety**: Process Manager uses mutex to prevent deadlocks (doneChan captured under lock, then released before waiting).
- **File Caching**: Plan/Specs views cache file content for 5 seconds to avoid I/O in render loop.

## Run Commands
- **Build**: `go build -o ralph-tui ./cmd/ralph-tui`
- **Run**: `./ralph-tui [--mode build|plan|plan-work] [--max N] [--work "desc"] [--script path/to/loop.sh]`
- **Test**: `go test ./...` (comprehensive tests for manager lifecycle, ring buffer, state)
