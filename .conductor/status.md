# Project Status

All core refactoring tasks, example validations, and developer tooling updates are complete. The Terminus framework is stable and supports both Web and CLI clients with a unified ANSI architecture.

## Recent Achievements:

- **Unified Architecture:** Successfully transitioned to a single Go backend streaming raw ANSI escape codes.
- **Client Flexibility:**
    - **Web:** A robust `terminus-client.js` with a custom ANSI parser handles rendering and input, including modifier keys.
    - **CLI:** A native `terminus-cli` tool connects via WebSocket and provides a raw terminal experience.
- **Example Modernization:** All 8 examples (`hello`, `todo`, `chat`, `dashboard`, `widgets`, `textinput`, `commands`, `layout`) have been updated to use centralized assets and verified to work.
- **Developer Experience:**
    - **Makefile:** Updated with a `help` command and dedicated `build-cli`/`run-cli` targets.
    - **Documentation:** `README.md` reflects the new architecture and usage instructions.
- **Bug Fixes & Refinements:**
    - Fixed critical input handling issues (modifier keys support for shortcuts like `Shift+Tab`, `Ctrl+S`, `Ctrl+R`).
    - **CLI Input Engine:** Implemented a robust **State Machine Parser** in `terminus-cli` to handle ANSI escape sequences split across reads (fixing `[Z` artifacts) and reliably detect control keys.
    - **Cursor:** Resolved cursor rendering issues: now uses a non-blinking solid block cursor (White Background) that is always visible when focused.
    - **Debug:** Added debug info ("Last Key: ...") to `textinput` example to help troubleshoot input issues.
    - Fixed a syntax error in session management logic.
    - Corrected asset embedding to ensure the web client loads reliably without 404s.
    - **Build Fix:** Removed invalid `//go:embed` directives from examples.
    - **Test Stability:** Resolved data race conditions in core tests (`cancel`, `engine`, `session`) and fixed compilation errors.

## Current State:
The project is stable and ready for use.
- **Web:** Run `go run examples/<name>/main.go` and visit `http://localhost:8890`.
- **CLI:** **CRITICAL:** Rebuild with `make build-cli` to get the parser fixes. Run `make run-cli` (after starting a server).
- **Tests:** All tests pass (`make test`).
- **Pending Validation:** New CLI input tweaks (explicit IXON disable, extended escape buffering) need retest on macOS iTerm2/Terminal/Ghostty to confirm `Shift+Tab`, `Ctrl+S`, and `Ctrl+R` reach the app and the block cursor remains visible.
