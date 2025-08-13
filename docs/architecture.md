# Terminus Architecture

This document describes the architecture and design decisions behind Terminus, a Go framework for building terminal-style user interfaces in web browsers.

## Overview

Terminus follows a Model-View-Update (MVU) architecture pattern, popularized by Elm. The framework consists of:

1. **Server-side Go application** - Handles all business logic and rendering
2. **WebSocket communication layer** - Real-time bidirectional communication
3. **Client-side JavaScript** - Thin client that displays content and captures input
4. **Component system** - Composable UI building blocks

```
┌─────────────────────┐         ┌─────────────────────┐
│   Browser Client    │         │    Go Server        │
│                     │         │                     │
│ ┌─────────────────┐ │         │ ┌─────────────────┐ │
│ │ JavaScript      │ │ WebSocket │ │  Component      │ │
│ │ - Input capture │◄├─────────┤►│  - Model        │ │
│ │ - DOM rendering │ │  JSON   │ │  - Update logic │ │
│ │ - ANSI parsing  │ │ Messages│ │  - View render  │ │
│ └─────────────────┘ │         │ └─────────────────┘ │
│                     │         │                     │
│ ┌─────────────────┐ │         │ ┌─────────────────┐ │
│ │ Terminal CSS    │ │         │ │ Session Manager │ │
│ │ - Styling       │ │         │ │ - Lifecycle     │ │
│ │ - Layout        │ │         │ │ - State         │ │
│ └─────────────────┘ │         │ └─────────────────┘ │
└─────────────────────┘         └─────────────────────┘
```

## Core Concepts

### Model-View-Update Pattern

The MVU pattern provides a predictable way to manage application state:

1. **Model** - The application state
2. **View** - A pure function that renders the model
3. **Update** - A pure function that updates the model based on messages

```go
type Component interface {
    Init() Cmd                    // Initialize component
    Update(Msg) (Component, Cmd)  // Handle messages
    View() string                 // Render to string
}
```

### Messages and Commands

- **Messages** (`Msg`) - Events that trigger state updates
- **Commands** (`Cmd`) - Side effects that produce messages

This separation ensures:
- Pure update functions
- Testable business logic
- Predictable state management

### Server-Side Rendering

All rendering happens on the server:

1. Component renders to ANSI-styled text
2. Diff algorithm calculates minimal changes
3. Updates sent to client as JSON commands
4. Client applies updates to DOM

Benefits:
- Zero client-side state management
- Consistent rendering across browsers
- Simplified security model
- Easy testing

## Component Architecture

### Component Lifecycle

```
┌──────────┐     ┌────────┐     ┌────────┐     ┌──────┐
│   Init   │────►│ Update │────►│  View  │────►│ Diff │
└──────────┘     └────┬───┘     └────────┘     └───┬──┘
                      │                             │
                      │      ┌──────────┐           │
                      └──────┤ Commands │◄──────────┘
                             └──────────┘
```

1. **Initialization** - `Init()` called once when component starts
2. **Message Loop** - `Update()` processes messages
3. **Rendering** - `View()` produces string output
4. **Diffing** - Changes calculated and sent to client

### Session Management

Each WebSocket connection gets its own session:

```go
type Session struct {
    ID        string
    conn      *websocket.Conn
    component Component
    engine    *Engine
    // ... other fields
}
```

Sessions are independent:
- Isolated state per user
- Concurrent execution
- Automatic cleanup on disconnect

## Networking Layer

### WebSocket Protocol

Messages use JSON for simplicity:

#### Client to Server:
```json
{
  "type": "key",
  "data": {
    "keyType": "enter"
  }
}
```

#### Server to Client:
```json
{
  "type": "render",
  "data": {
    "content": "Hello, \u001b[31mWorld\u001b[0m!"
  }
}
```

### Message Types

**Client Messages:**
- `key` - Keyboard input
- `resize` - Terminal dimensions
- `mouse` - Mouse events (future)

**Server Messages:**
- `render` - Full screen render
- `updateLine` - Single line update
- `clear` - Clear screen
- `batch` - Multiple commands

## Rendering Pipeline

### 1. Style System

Fluent API for text styling:

```go
style.New().
    Foreground(style.Red).
    Bold(true).
    Render("Hello")
// Output: "\x1b[31;1mHello\x1b[0m"
```

### 2. Virtual Screen

Line-based buffer for efficient updates:

```go
type Screen struct {
    lines  []string
    width  int
    height int
}
```

### 3. Diff Algorithm

Calculates minimal updates between renders:

```go
type DiffCommand struct {
    Type    string // "updateLine", "clear", etc.
    Content string
    Line    int
}
```

### 4. Client Rendering

JavaScript parses ANSI and updates DOM:

```javascript
class ANSIParser {
    parse(text) {
        // Convert ANSI escape sequences to HTML
        // Handle colors, styles, cursor positioning
    }
}
```

## Widget System

### Widget Interface

Extends Component with focus management:

```go
type Widget interface {
    Component
    Focus()
    Blur()
    Focused() bool
}
```

### Base Widget Model

Common functionality for all widgets:

```go
type Model struct {
    focused  bool
    disabled bool
    x, y     int
    width    int
    height   int
}
```

### Widget Types

1. **Input Widgets**
   - TextInput - Single-line text entry
   - TextArea - Multi-line text (planned)
   - Select - Dropdown selection (planned)

2. **Display Widgets**
   - List - Scrollable item list
   - Table - Data grid with sorting
   - Tree - Hierarchical data (planned)

3. **Feedback Widgets**
   - Spinner - Loading animations
   - Progress - Progress bars (planned)
   - Toast - Notifications (planned)

## Performance Considerations

### Server-Side Optimizations

1. **Diff Algorithm** - O(n) line-based comparison
2. **Render Throttling** - Batch updates within frame
3. **Command Pooling** - Reuse command objects
4. **Goroutine Pool** - Limited concurrent commands

### Client-Side Optimizations

1. **DOM Patching** - Update only changed lines
2. **ANSI Caching** - Cache parsed styles
3. **Debounced Resize** - Limit resize events
4. **Virtual Scrolling** - Render visible content only

### Network Optimizations

1. **Message Compression** - Gzip WebSocket frames
2. **Delta Updates** - Send only changes
3. **Batch Commands** - Group related updates
4. **Binary Protocol** - Future optimization

## Security Model

### Input Validation

All input sanitized server-side:
- HTML escaping
- ANSI sequence validation
- Command injection prevention

### Session Isolation

- No shared state between sessions
- Separate goroutines per session
- Resource limits per connection

### WebSocket Security

- Origin validation
- Rate limiting
- Connection timeouts
- Message size limits

## Extension Points

### Custom Components

Easy to create custom components:

```go
type MyComponent struct {
    model MyModel
}

func (m *MyComponent) Init() terminus.Cmd { 
    return nil 
}

func (m *MyComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    // Handle messages
}

func (m *MyComponent) View() string {
    // Render view
}
```

### Custom Commands

Create async operations:

```go
func FetchData(url string) terminus.Cmd {
    return func() terminus.Msg {
        // Perform async operation
        resp, err := http.Get(url)
        return DataMsg{resp, err}
    }
}
```

### Custom Widgets

Extend base widget functionality:

```go
type MyWidget struct {
    widget.Model
    // Custom fields
}

// Implement Widget interface
```

## Testing Strategy

### Unit Testing

Components are pure functions:

```go
func TestComponent(t *testing.T) {
    comp := NewComponent()
    
    // Test initialization
    cmd := comp.Init()
    assert.Nil(t, cmd)
    
    // Test update
    comp, cmd = comp.Update(KeyMsg{Type: KeyEnter})
    assert.Equal(t, expected, comp.View())
}
```

### Integration Testing

Test full session lifecycle:

```go
func TestSession(t *testing.T) {
    session := NewSession(mockConn, NewComponent())
    
    // Send message
    session.HandleMessage(KeyMsg{})
    
    // Verify output
    assert.Contains(t, mockConn.Written(), expected)
}
```

### E2E Testing

Browser automation with WebSocket mocking:

```javascript
describe('Terminus App', () => {
    it('handles keyboard input', () => {
        // Mock WebSocket
        // Send key events
        // Verify DOM updates
    });
});
```

## Future Enhancements

### Planned Features

1. **Mouse Support** - Click, drag, hover events
2. **File Upload** - Drag-and-drop files
3. **Audio** - Terminal bell, notifications
4. **Themes** - Customizable color schemes
5. **Plugins** - Extension system

### Performance Improvements

1. **Binary Protocol** - Replace JSON with MessagePack
2. **Compression** - Built-in compression
3. **Caching** - Client-side render cache
4. **Web Workers** - Offload parsing

### Developer Experience

1. **Hot Reload** - Live component updates
2. **DevTools** - Browser extension
3. **Playground** - Online editor
4. **Generator** - Project scaffolding

## Design Decisions

### Why Server-Side Rendering?

1. **Simplicity** - No client state synchronization
2. **Security** - All logic server-side
3. **Consistency** - Same render everywhere
4. **Performance** - Minimal client code

### Why Line-Based Diff?

1. **Efficiency** - Natural terminal boundary
2. **Simplicity** - Easy to implement
3. **Performance** - O(n) complexity
4. **Compatibility** - Works with ANSI

### Why Go?

1. **Performance** - Fast execution
2. **Concurrency** - Built-in primitives
3. **Simplicity** - Easy to learn
4. **Deployment** - Single binary

### Why MVU?

1. **Predictability** - Pure functions
2. **Testability** - Easy to test
3. **Debugging** - Time-travel debugging
4. **Simplicity** - One-way data flow

## Conclusion

Terminus provides a unique approach to building web-based terminal UIs by combining:

- Server-side rendering for simplicity
- MVU pattern for predictability
- WebSocket for real-time updates
- ANSI styling for rich text

This architecture enables developers to build sophisticated terminal applications that run in any modern browser while maintaining the simplicity and power of terminal interfaces.