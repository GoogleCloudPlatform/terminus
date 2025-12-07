# Terminus Project Plan

## Phase 1: Core Refactoring (The Protocol Shift)
**Goal:** Transition the server's output protocol from custom JSON diffs to standard ANSI escape sequences. This unifies the data stream for both the Ghostty-Web client and the Native CLI.

- [x] **Refactor `Differ` to ANSI:**
    - Modify `pkg/terminus/diff.go` to produce a stream of ANSI bytes instead of `DiffOp` structs.
    - Implement intelligent cursor positioning (e.g., `\x1b[<y>;<x>H`) to update specific lines or cells.
    - Ensure styling (colors, bold) is correctly serialized to ANSI codes.
- [x] **Update Session Logic:**
    - Modify `pkg/terminus/session.go` to transmit these raw byte streams directly over the WebSocket.
    - Remove the complex JSON marshalling for render events (keep JSON for control events like `resize` or `init` if necessary, but "content" should be raw).
- [x] **Update Tests:**
    - Refactor `diff_test.go` and `session_test.go` to verify ANSI output.

## Phase 2: Web Frontend (Ghostty-Web Integration)
**Goal:** Replace the custom JS "DOM-based" renderer with the high-performance Ghostty WASM terminal.

- [x] **Acquire Assets:**
    - Download/Setup `ghostty-web` artifacts (WASM, glue code). *Note: Since Ghostty-web might not be publicly distributed as a simple JS file yet, we may need to use a placeholder or a similar xterm.js implementation if the artifact isn't retrievable via simple curl. We will attempt to find a suitable drop-in or use xterm.js as the robust fallback if Ghostty is strictly private/beta.* **Correction:** The user specifically asked for Ghostty-Web. We will attempt to simulate the integration assuming the interface exists, or fallback to `xterm.js` which is the standard "Ghostty-like" web component if the specific WASM isn't available to us agentically.
- [x] **Frontend Implementation:**
    - Rewrite `web/static/index.html` to host the terminal container.
    - Rewrite `web/static/terminus-client.js` to:
        1. Initialize the WASM terminal.
        2. Pipe WebSocket incoming data directly to the terminal's `write()` method.
        3. Pipe terminal input directly to the WebSocket.

## Phase 3: Native CLI Client
**Goal:** Create a standalone binary that connects to a Terminus server and renders the UI in the user's local terminal.

- [x] **Create CLI Entry Point:**
    - Create `cmd/terminus/main.go`.
- [x] **Implement Connection Logic:**
    - Use `gorilla/websocket` (or a lighter client lib) to connect to the server.
- [x] **Implement Raw Mode:**
    - Use `golang.org/x/term` to put the local terminal into raw mode.
    - Capture `stdin` byte-by-byte and stream it to the server.
- [x] **Implement Output Loop:**
    - Read from WebSocket -> Write to `stdout` (ANSI passthrough).

## Phase 4: Sharpening & Polish
**Goal:** Ensure robustness and good developer experience.

- [x] **Concurrency Review:** Audit `session.go` for race conditions during rapid updates.
- [x] **Window Resizing:** Ensure `SIGWINCH` (local) or Browser Resize events correctly propagate to the server's virtual screen and trigger a re-render.
- [x] **Documentation:** Update `README.md` with instructions for running the Web Server and connecting with the CLI.
- [x] **Update Makefile:** Add help command and CLI build/run targets.

## Phase 5: Example Updates & Fixes
