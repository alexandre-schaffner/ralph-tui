# Architectural Context

## Project Pattern
- **Wrapper TUI**: This application is a terminal user interface wrapper around an existing bash script (`loop.sh`).
- **State Management**: Uses a centralized state store (Redux-like or simple struct) passed to Bubbletea models.
- **Concurrency**: `loop.sh` runs in a separate goroutine; stdout/stderr are streamed to the UI via channels or a thread-safe buffer.

## Key Gotchas
- **Process Control**: Killing the TUI must ensure the child process (`loop.sh`) is also terminated.
- **Streaming**: Large output from `loop.sh` needs efficient buffering to avoid UI lag.
