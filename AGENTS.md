# Architectural Context

## Project Pattern
- **Wrapper TUI**: This application is a terminal user interface wrapper around an existing bash script (`loop.sh`).
- **State Management**: Uses a centralized state store (`src/lib/state`) with thread-safe access passed to Bubbletea models.
- **Concurrency**: `loop.sh` runs in a subprocess; stdout/stderr are streamed to a ring buffer via goroutines.

## Key Gotchas
- **Process Control**: Killing the TUI must ensure the child process (`loop.sh`) is also terminated (SIGTERM → 5s → SIGKILL).
- **Streaming**: Large output from `loop.sh` uses a 1000-line ring buffer to prevent UI lag.

## Run Commands
- **Build**: `go build -o ralph-tui ./cmd/ralph-tui`
- **Run**: `./ralph-tui [--mode build|plan|plan-work] [--max N] [--work "desc"]`
- **Test**: `go test ./...`
