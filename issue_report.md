# Bug Report: CLI Input Handling (Ctrl+S, Shift+Tab, Cursor)

## 1. Issue Description

The user is reporting persistent issues with the `terminus-cli` native client when interacting with the `textinput` example application. Specifically:

### A. Input Handling Failures
1.  **Shift+Tab:** 
    -   **Expected:** Should navigate to the previous field.
    -   **Actual:** Renders as `[Z` in the input field. The application debug info shows "Last Key: Z" (or similar).
    -   **Interpretation:** The ANSI escape sequence for Shift+Tab (`\x1b[Z`) is not being parsed correctly as a single unit. It appears to be splitting into `\x1b` (Escape, possibly ignored or consumed) + `[` (Bracket literal) + `Z` (Letter Z). The application receives the literal characters.

2.  **Ctrl+S:**
    -   **Expected:** Should submit the form (application logic).
    -   **Actual:** Nothing happens. The application debug info ("Last Key: ...") shows *nothing* changing when `Ctrl+S` is pressed.
    -   **Interpretation:** The keystroke `Ctrl+S` (ASCII 19, `\x13`) is not reaching the application logic. This strongly suggests it is being intercepted by the system terminal driver (XOFF flow control) despite the CLI attempting to enter raw mode.

3.  **Ctrl+R:**
    -   **Expected:** Should reset the form.
    -   **Actual:** Reported as "not working". Likely similar to `Ctrl+S` or input mapping issue.

### B. Visual Artifacts
1.  **Cursor Appearance:**
    -   **Initially:** User reported a "white vertical bar" at launch (likely the default pipe character `|`).
    -   **Update:** User requested a "white square" (block cursor) that is "always visible" and "non-blinking".
    -   **Status:** We attempted to fix this by changing the style to `Background(White)` and character to space `' '`, and disabling blinking.

## 2. Files Involved

### Core CLI Client
-   **`cmd/terminus/main.go`**: The entry point for the native CLI.
    -   **Responsibility:** Puts terminal in raw mode, reads `os.Stdin`, parses bytes/sequences, marshals them into JSON `ClientMessage`, and sends them over WebSocket.
    -   **Current State:** Contains a custom loop reading into a 1024-byte buffer with manual lookahead parsing for escape sequences (`\x1b...`) and control keys.

### Server-Side Session Handling
-   **`pkg/terminus/session.go`**: Handles WebSocket connections on the server.
    -   **Responsibility:** Receives `ClientMessage` JSON, parses `keyType` and `modifiers`, converts them to internal `KeyMsg` structs, and sends them to the `Engine`.
    -   **Current State:** Updated to parse `modifiers` map (ctrl, shift, alt) and handle `ctrl+s`/`ctrl+r` string types.

### Widget Implementation
-   **`pkg/terminus/widget/textinput.go`**: The text input component.
    -   **Responsibility:** Renders the input field and cursor, handles `KeyMsg` events.
    -   **Current State:** Modified to use solid block cursor, disabled blinking, and restored missing configuration methods.

### Example Application
-   **`examples/textinput/main.go`**: The demo app.
    -   **Responsibility:** Sets up the UI and handles application-specific logic (submit on `KeyCtrlS`, reset on `KeyCtrlR`).
    -   **Current State:** Added debug output to show the last received key.

## 3. Attempted Fixes

### Fix 1: Protocol & Mapping (Web & CLI)
-   **Action:** Updated `session.go` to support `modifiers` field in JSON. Updated `terminus-client.js` and `terminus-cli` to send this field. Added specific cases for `Ctrl+S` and `Ctrl+R`.
-   **Result:** Web client (presumably) works better (preventDefault added). CLI still failing.

### Fix 2: Cursor Styling
-   **Action:** In `textinput.go`, changed `cursorChar` from `|` to ` ` (space), set style to `Background(White)`, and removed `blink` logic.
-   **Result:** Should show a solid block.

### Fix 3: CLI Parser Rewrite (Attempt to fix Shift+Tab)
-   **Action:** In `cmd/terminus/main.go`, replaced simple `string(buf) == "\x1b[Z"` check with a loop that iterates through the buffer (`buf[:n]`) byte-by-byte but looks ahead when `\x1b` is encountered.
    -   Logic: If `buf[i] == '\x1b'`, checks `buf[i+1]` and `buf[i+2]`. If match `[Z`, consumes 3 bytes and sends `KeyTab` + `Shift`.
-   **Result:** User reports it *still* renders `[Z`. This implies the lookahead failed or the bytes arrived in separate `Read` calls (e.g. split packet), or `term.MakeRaw` behavior on the user's system is unexpected.

### Fix 4: Ctrl+S Debugging
-   **Action:** Added extensive logging in `cmd/terminus/main.go` (`log.Printf`) to trace every byte read.
-   **Hypothesis:** `Ctrl+S` is XOFF. `term.MakeRaw` *should* disable it (`IXON`), but if it's not working, `Read` will block/pause or simply not receive the byte.

## 4. Next Steps / Investigation

1.  **Shift+Tab Persistence:**
    -   If `Read` returns `\x1b` in one call, and `[Z` in the next call, the current lookahead logic (which only looks at `n` bytes read *currently*) will fail. It treats the isolated `\x1b` as "Escape" key, sends it, then processes `[` and `Z` as text.
    -   **Solution:** We need a **stateful parser** that buffers incomplete escape sequences across `Read` calls.

2.  **Ctrl+S Persistence:**
    -   If `MakeRaw` is working, `Ctrl+S` should be byte 19.
    -   If it's not appearing, something lower level is eating it.
    -   **Verification:** We need to confirm if the user's terminal is actually entering raw mode correctly.
