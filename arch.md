Okay, let's dive deep into a detailed architecture for **TerminusGo**. This will cover the server-side Go components, the client-side JavaScript, the communication protocol, and how they all fit together.

## TerminusGo: Detailed Architecture

**Guiding Principles:**

*   **Go-centric:** Most logic resides in Go.
*   **Model-View-Update (MVU) Core:** The primary architectural pattern for Go components.
*   **Thin Client:** The browser client is primarily for rendering, input forwarding, and basic WebSocket management.
*   **WebSocket Communication:** The backbone for real-time, bidirectional communication.
*   **Optimized for Text:** The rendering pipeline is designed for efficient transmission and display of styled text.
*   **Extensible:** Allow for custom components, commands, and potentially client-side render plugins.

**High-Level Diagram:**

```
+--------------------------------------------------------------------------------------+
| User's Web Browser                                                                   |
| +-----------------------------+      WebSocket      +--------------------------------+ |
| | Client-Side JavaScript      |<------------------->| TerminusGo Server (Go App)     | |
| | (terminus-client.js)        |      (JSON/Binary)  | +----------------------------+ | |
| |                             |                     | | Program                    | | |
| | - WebSocket Manager         |                     | |  - Session Manager         | | |
| | - Input Handler (Keyboard)  |                     | |  - HTTP Handler            | | |
| | - Renderer (Virtual Text   |                     | |  - WebSocket Upgrader      | | |
| |   Screen/DOM Patcher)       |                     | | +------------------------+ | | |
| | - Style Parser              |                     | | | Session (Per Client)   | | | |
| +-----------------------------+                     | | |  - Root Component      | | | |
|                                                     | | |  - Command Processor   | | | |
|                                                     | | |  - Message Queue       | | | |
|                                                     | | |  - Renderer (Go Side)  | | | |
|                                                     | | +------------------------+ | | |
|                                                     | +----------------------------+ | |
|                                                     +--------------------------------+ |
+--------------------------------------------------------------------------------------+
```

---

**I. Server-Side (Go Application)**

**A. `Program` (The Core Orchestrator)**

*   **Responsibilities:**
    *   Entry point for a TerminusGo application.
    *   Manages HTTP server for serving the initial HTML/JS client and upgrading to WebSockets.
    *   Manages active client sessions.
    *   Handles graceful shutdown.
*   **Key Components:**
    *   **`HTTPServer`:** Standard Go `http.Server`.
    *   **`HTTPHandler`:**
        *   Serves `index.html` and `terminus-client.js`.
        *   Handles WebSocket upgrade requests on a specific endpoint (e.g., `/ws`).
    *   **`SessionManager`:**
        *   Map of active sessions (e.g., `map[sessionID]*Session`).
        *   Creates new sessions on WebSocket connection, destroys them on disconnection.
    *   **`RootComponentFactory func() terminus.Component`:** A function provided by the user that creates an instance of their root UI component.

**B. `Session` (Per Connected Client)**

*   **Responsibilities:**
    *   Represents a single connected client.
    *   Owns the instance of the user's root `terminus.Component`.
    *   Manages the MVU loop for that client.
    *   Processes outgoing render commands and incoming client messages.
*   **Key Components:**
    *   **`WebSocketConnection *websocket.Conn`:** The actual WebSocket connection.
    *   **`RootComponent terminus.Component`:** The instance of the user's UI.
    *   **`MessageQueue chan terminus.Msg`:** An internal channel to serialize incoming messages (from client input or command results) before processing by `Update()`.
    *   **`CommandProcessor`:**
        *   Receives `terminus.Cmd` functions from `Update()`.
        *   Executes these commands (often in separate goroutines for non-blocking I/O).
        *   Sends the resulting `terminus.Msg` back to the `MessageQueue`.
    *   **`Renderer (Go Side)`:**
        *   After `Update()`, calls `RootComponent.View()`.
        *   Takes the returned string (which may contain style escape codes).
        *   **Diffing (Crucial Optimization):**
            *   Maintains the "previous" rendered string or a structured representation.
            *   Compares the new view string with the previous one.
            *   Computes a diff (e.g., line-based diff, or more granular if a virtual grid is used).
            *   Sends only the *changes* to the client, not the full screen on every update. This is critical for performance.
    *   **`outgoingMessages chan []byte`:** Channel for messages to be sent to the WebSocket client.
*   **MVU Loop within a Session:**
    1.  `Init()` called on `RootComponent` when session starts. Any `Cmd` is processed.
    2.  **Listen Loop (for WebSocket and Command results):**
        *   Receives client input messages (e.g., `KeyMsg`) from WebSocket. Pushes to `MessageQueue`.
        *   Receives command result messages from `CommandProcessor`. Pushes to `MessageQueue`.
    3.  **Update Loop (triggered by messages in `MessageQueue`):**
        *   Pops `terminus.Msg` from `MessageQueue`.
        *   Calls `newComponent, cmd = RootComponent.Update(msg)`.
        *   Updates `RootComponent = newComponent`.
        *   If `cmd` is not nil, sends it to `CommandProcessor`.
        *   If `cmd == terminus.Quit`, initiates session shutdown.
        *   Triggers a re-render:
            *   `viewString = RootComponent.View()`.
            *   `diffCommands = Renderer.DiffAndPackage(previousViewString, viewString)`.
            *   Sends `diffCommands` over WebSocket to the client.
            *   `previousViewString = viewString`.

**C. `terminus.Component` Interface**

```go
package terminus

type Msg interface{} // Marker interface for messages

type Cmd func() Msg // A command is a function that produces a message (or nil)

// Special command to signal quitting
var Quit Cmd = func() Msg { return quitMsg{} } // quitMsg is an internal type

type KeyType int
const (
    KeyRunes KeyType = iota // Character input
    KeyEnter
    KeySpace
    KeyBackspace
    // ... other keys like Tab, Esc, Arrows, F1-F12, Ctrl+C, etc.
)

type KeyMsg struct {
    Type  KeyType
    Runes []rune // For KeyRunes
    // Modifiers (Ctrl, Alt, Shift) could be added here
}

func (km KeyMsg) String() string { /* ... helper to get 'q', 'ctrl+c' etc. ... */ }


type Component interface {
    Init() Cmd
    Update(msg Msg) (Component, Cmd) // Returns the NEW state of the component and a command
    View() string                    // Returns the string representation to render
}
```

**D. Styling (`terminus/style` package)**

*   **Responsibility:** Provide a fluent API to define text styles (color, bold, italic, etc.).
*   **Mechanism:**
    *   The `style.New().Bold(true).Foreground(style.Color("red")).Render("text")` would produce a string with embedded, non-printable ANSI-like escape codes or custom control sequences.
    *   Example: `"\x1b[1m\x1b[31mtext\x1b[0m"` (ANSI) or a custom format like `§{fg:red,bold:true}text§{reset}`.
    *   The client-side renderer will parse these sequences. Using custom sequences might be easier to parse robustly than full ANSI, and allows defining only what TerminusGo supports.

**E. Diffing Algorithm (Server-Side Renderer)**

*   **Challenge:** Efficiently find differences between two multi-line strings representing terminal screens.
*   **Approaches:**
    1.  **Line-based Diff:**
        *   Split old and new view strings into lines.
        *   Use a standard diff algorithm (like Myers diff) to find differing, added, or removed lines.
        *   Commands sent to client: `UPDATE_LINE(idx, content)`, `INSERT_LINE(idx, content)`, `DELETE_LINE(idx)`.
    2.  **Virtual Grid Diff (More Complex, More Granular):**
        *   Server maintains a 2D grid of characters and their styles (a "virtual screen buffer").
        *   `View()` conceptually "paints" onto this virtual grid.
        *   Diffing compares the old grid state with the new grid state cell by cell.
        *   Commands sent to client: `SET_CELL(row, col, char, style)`, `CLEAR_RECT(x,y,w,h)`.
        *   This can be more efficient for small, localized changes but adds server-side complexity.
*   The choice depends on the desired granularity of updates and implementation complexity. Line-based is a good start.

---

**II. Client-Side (JavaScript: `terminus-client.js`)**

*   **Responsibilities:**
    *   Establish and maintain WebSocket connection.
    *   Capture keyboard input and send it to the server.
    *   Receive render commands from the server.
    *   Parse style sequences.
    *   Render the styled text to the browser (e.g., into a `<pre>` tag or a series of `<div>`s).
    *   Handle window resize events (inform server so `View()` can adapt).

*   **Key Modules/Components:**
    *   **`WebSocketManager`:**
        *   Connects to the `/ws` endpoint.
        *   Handles `onopen`, `onmessage`, `onerror`, `onclose` events.
        *   Reconnect logic (optional, but good for resilience).
        *   Sends serialized input events to the server.
    *   **`InputHandler`:**
        *   Attaches event listeners for `keydown`, `keypress` (or `beforeinput`).
        *   Translates browser key events into `terminus.KeyMsg`-like JSON objects.
        *   Example: `{ "type": "runes", "runes": ["h", "e", "l", "l", "o"] }` or `{ "type": "enter" }`.
        *   Sends these objects via `WebSocketManager`.
        *   Handles `preventDefault()` for keys that should be captured (e.g., Tab, Arrows).
    *   **`Renderer (Client-Side)` / `DOMPatcher`:**
        *   Receives render commands from the server (e.g., `UPDATE_LINE`, `SET_CELL`).
        *   Maintains a representation of the current screen state (e.g., an array of strings for lines, or a more complex structure if using a grid).
        *   **Rendering Target:**
            *   **Simple `<pre>` tag:** Easiest. Update `innerHTML` or `textContent`. Styles applied via `<span>` tags with inline styles or CSS classes.
            *   **Div-per-line / Span-per-char:** More complex but allows finer-grained DOM manipulation. Each line is a `<div>`, each character (or styled segment) is a `<span>`.
        *   **Style Parser:**
            *   Parses the style sequences (e.g., `§{fg:red,bold:true}text§{reset}`) from the server.
            *   Translates them into CSS styles applied to `<span>` elements.
            *   Example: `§{fg:red}` becomes `<span style="color: red;">`.
        *   Applies the received diff commands to update the DOM efficiently.
    *   **`ResizeHandler`:**
        *   Listens to browser `resize` events.
        *   Sends a message to the server (e.g., `{ "type": "resize", "cols": 80, "rows": 24 }`) indicating the new terminal dimensions (calculated based on font size and container dimensions).
        *   The server can then use this information in `View()` to re-layout content.

---

**III. Communication Protocol (WebSocket)**

*   **Format:** JSON is easiest to start with. Binary formats (like Protocol Buffers or MessagePack) can be more efficient for high-frequency updates but add complexity.
*   **Messages (Server to Client):**
    *   Render commands (depends on diffing strategy):
        *   `{ "action": "CLEAR_SCREEN" }`
        *   `{ "action": "UPDATE_LINE", "lineIndex": 5, "content": "Styled §{fg:blue}text§{reset}" }`
        *   `{ "action": "SET_CURSOR", "row": 10, "col": 5 }` (if cursor management is explicit)
        *   `{ "action": "BATCH_UPDATE", "commands": [ ...array of other commands... ] }`
    *   Special commands:
        *   `{ "action": "SET_TITLE", "title": "My App" }`
        *   `{ "action": "QUIT_ACK" }` (server acknowledges client can close)

*   **Messages (Client to Server):**
    *   Input events:
        *   `{ "type": "key", "keyType": "runes", "runes": ["a"], "modifiers": {"ctrl": false, ...} }`
        *   `{ "type": "key", "keyType": "enter", "modifiers": {} }`
    *   Resize events:
        *   `{ "type": "resize", "cols": 80, "rows": 25 }`
    *   Ping/Pong for keep-alive (handled by WebSocket libraries usually).

---

**IV. Workflow Example: User Types a Character**

1.  **Client:** User presses 'a'.
    *   `InputHandler` captures `keydown`.
    *   Creates JSON: `{ "type": "key", "keyType": "runes", "runes": ["a"] }`.
    *   `WebSocketManager` sends this JSON to the server.
2.  **Server (`Session`):**
    *   WebSocket connection receives the JSON.
    *   Deserializes it into a Go `terminus.KeyMsg{Type: terminus.KeyRunes, Runes: []rune{'a'}}`.
    *   This `KeyMsg` is put onto the `MessageQueue` for the session.
3.  **Server (MVU Loop in `Session`):**
    *   The `KeyMsg` is picked from the `MessageQueue`.
    *   `newComp, cmd = RootComponent.Update(keyMsg)` is called.
        *   Let's say `Update` appends 'a' to a model's string field.
    *   `RootComponent` is updated to `newComp`.
    *   `cmd` (if any) is sent to `CommandProcessor`.
    *   `viewStr = RootComponent.View()` is called.
    *   `diffCmds = Renderer.DiffAndPackage(prevViewStr, viewStr)` is called.
        *   This might result in: `{ "action": "UPDATE_LINE", "lineIndex": 3, "content": "Current text: a" }` (if 'a' was added to line 3).
    *   This diff command (JSON) is sent via WebSocket to the client.
4.  **Client:**
    *   `WebSocketManager` receives the diff command JSON.
    *   `Renderer` parses it.
    *   `DOMPatcher` updates the appropriate part of the DOM. For example, it finds the `<span>` or `<div>` for line 3 and updates its content, applying any new styles.
5.  The user sees the character 'a' appear on the screen.

---

**Key Architectural Decisions & Trade-offs:**

*   **Diffing Strategy:** Line-based vs. grid-based. Impacts complexity and performance.
*   **Style Encoding:** ANSI vs. custom sequences. Custom is more controlled but less standard.
*   **Client-Side Rendering Target:** `<pre>` vs. more structured DOM. `<pre>` is simpler; structured DOM offers more control but is harder to manage for text layout.
*   **State Management in Go Components:** Pure MVU vs. allowing mutable state within component structs (MVU typically encourages immutability by returning new model instances).
*   **Command Execution:** Synchronous vs. asynchronous for `Cmd` functions. Asynchronous is generally better for non-blocking I/O.

This detailed architecture provides a strong foundation for TerminusGo. It's ambitious but creates a clear separation of concerns and a robust pipeline for delivering interactive text UIs to the browser using Go. Remember, this is a "for fun" project, so start with the simplest viable version of each component and iterate!
