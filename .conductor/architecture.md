# Terminus Architecture

## 1. System Overview

Terminus is a framework for building terminal-style applications that run ubiquitously. It decouples the application logic from the presentation layer, allowing the same Go code to power interfaces in a web browser and a native terminal.

It adheres to the **Model-View-Update (MVU)** architecture, ensuring a unidirectional data flow and strict state management.

The system consists of three primary components:
1.  **The Server (Go):** Executes application logic, maintains state, and renders the UI to a virtual screen.
2.  **The Browser Client (Ghostty-Web):** A WASM-powered rendering layer for the web.
3.  **The Native Client (CLI):** A lightweight local executable acting as a bridge for standard terminal emulators.

```mermaid
graph TD
    subgraph "Clients"
        Browser[Browser (Ghostty-Web)] -->|JSON/WS| Server
        CLI[Native Terminal CLI] -->|JSON/WS| Server
    end
    
    subgraph "Server (Go)"
        Server -->|Update Model| App[Application Logic]
        App -->|Render View| VirtualScreen[Virtual Screen]
        VirtualScreen -->|Diff/ANSI| Server
    end
    
    Server -->|Render Commands| Browser
    Server -->|Render Commands| CLI
```

## 2. Core Concepts

### 2.1. The Component Model (MVU)
Every UI element in Terminus is a `Component` implementing this interface:

```go
type Component interface {
    // Init initializes the component and optionally returns a command.
    Init() Cmd

    // Update handles a message and returns an updated model and a command.
    Update(Msg) (Component, Cmd)

    // View renders the component's state as a string (with ANSI codes).
    View() string
}
```

-   **Model:** The immutable state of the application.
-   **View:** A pure function transforming the Model into a visual representation (string).
-   **Update:** A pure function transforming the Model based on a Message.

### 2.2. The Engine
The `Engine` manages the lifecycle of a user session.
1.  It holds the active `Component`.
2.  It maintains an event loop processing `Msg`s.
3.  After every update, it calls `View()`.
4.  It calculates the difference between the previous and current frame (`ScreenDiffer`) to minimize network traffic.

## 3. Communication Protocol

Communication occurs over **WebSockets**. The protocol is message-based using JSON, agnostic to the client type.

### 3.1. Client-to-Server
The client sends events captured by the terminal emulator or browser window.

```json
{
  "type": "key",
  "data": {
    "keyType": "char",
    "runes": ["a"],
    "modifiers": ["ctrl"]
  }
}
```

Supported message types:
-   `key`: Keyboard input.
-   `resize`: Terminal dimension changes.
-   `mouse`: Mouse interactions (future).

### 3.2. Server-to-Client
The server sends rendering commands or control signals.

```json
{
  "type": "render", // or "updateLine", "setCell", "clear"
  "data": {
    "content": "\x1b[31mHello World\x1b[0m", // ANSI string
    "y": 5
  }
}
```

## 4. Rendering Pipeline

### 4.1. Server-Side Rendering
The `View()` function produces a standard string containing ANSI escape codes for styling (colors, bold, etc.). The server uses a virtual screen buffer to track the current state of the terminal.

### 4.2. Diffing Strategy
To ensure performance:
1.  The engine compares the new frame against the previous frame.
2.  It generates a list of operations (e.g., "redraw line 5", "clear screen").
3.  These operations are batched and sent to the client.

### 4.3. Client Implementations

#### A. Browser (Ghostty-Web)
Uses **ghostty-web**, a high-performance terminal emulator compiled to WebAssembly.
1.  **Input:** Receives JSON commands.
2.  **Output:** Writes ANSI payloads directly to the `ghostty` instance (WebGL/Canvas).

#### B. Native CLI
A simple Go executable that users install locally.
1.  **Connect:** Establishes a WebSocket connection to the Terminus server.
2.  **Input:** Sets the local terminal to "raw mode," captures stdin, and marshals events into JSON.
3.  **Output:** Unmarshals JSON render commands and writes the raw ANSI content to stdout.

## 5. Session Management

The `SessionManager` handles multiple concurrent WebSocket connections.
-   **Isolation:** Each connection spawns a new goroutine and a dedicated `Engine` instance.
-   **State:** User state is kept in memory on the server for the duration of the connection.
-   **Lifecycle:** Disconnection triggers context cancellation and cleanup of the associated Engine.

## 6. Security & Validation

-   **Input Sanitization:** All incoming JSON is unmarshaled into strict types.
-   **Resource Limits:** The `Session` enforces read/write deadlines and buffer sizes to prevent DoS attacks.
-   **Sandbox:** The server logic is isolated from the client; the client can only send predefined event messages.
