# Terminus

A Go framework for building terminal-style user interfaces that run in web browsers. Terminus brings the simplicity and power of terminal UIs to the web, using a Model-View-Update (MVU) architecture similar to Elm.

## ✨ Features

- **🖥️ Terminal-Style UI in the Browser**: Build text-based interfaces that run in any modern web browser
- **🎯 MVU Architecture**: Simple, predictable state management with Model-View-Update pattern
- **📦 Rich Widget Library**: Pre-built components for common UI needs
- **🚀 Server-Side Rendering**: Keep your application logic in Go, minimize client-side JavaScript
- **⚡ Real-Time Updates**: WebSocket-based communication for instant UI updates
- **🎨 Full Styling Support**: Colors, bold, italic, underline, and more using ANSI-style formatting
- **📱 Responsive Design**: Works on desktop and mobile browsers
- **🔧 Easy to Extend**: Simple component interface for building custom widgets

## Widget Library

### TextInput
Full-featured text input with:
- Cursor movement and text editing
- Validation support
- Placeholder text
- Max length constraints
- Custom styling for all states

### List
Scrollable lists with:
- Keyboard navigation
- Real-time filtering
- Custom item rendering
- Selection callbacks
- Wrap-around navigation

### Table
Data tables with:
- Sortable columns
- Cell/row selection
- Custom alignment
- Row numbers
- Header customization

### Spinner
Loading indicators with:
- Multiple animation styles
- Custom characters
- Configurable speed
- Text positioning

## 🚀 Quick Start

### Installation

```bash
go get github.com/GoogleCloudPlatform/terminus
```

### Hello World Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/GoogleCloudPlatform/terminus/pkg/terminus"
    "github.com/GoogleCloudPlatform/terminus/terminus/style"
)

type HelloWorld struct {
    count int
}

func (h *HelloWorld) Init() terminus.Cmd {
    return nil
}

func (h *HelloWorld) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    switch msg := msg.(type) {
    case terminus.KeyMsg:
        switch msg.Type {
        case terminus.KeyUp:
            h.count++
        case terminus.KeyDown:
            h.count--
        case terminus.KeyEscape:
            return h, terminus.Quit
        }
    }
    return h, nil
}

func (h *HelloWorld) View() string {
    title := style.New().Bold(true).Foreground(style.Cyan).Render("Terminus Counter")
    counter := style.New().Bold(true).Render(fmt.Sprintf("Count: %d", h.count))
    help := style.New().Faint(true).Render("↑/↓ to change • ESC to quit")

    return fmt.Sprintf("%s\n\n%s\n\n%s", title, counter, help)
}

func main() {
    program := terminus.NewProgram(
        func() terminus.Component { return &HelloWorld{} },
        terminus.WithAddress(":8080"),
    )

    fmt.Println("Starting on http://localhost:8080")
    if err := program.Start(); err != nil {
        log.Fatal(err)
    }
    program.Wait()
}
```

## 📚 Documentation

- [**Getting Started**](docs/getting-started.md) - Quick introduction and your first app
- [**Tutorial**](docs/tutorial.md) - Build a complete todo list application
- [**API Reference**](docs/api.md) - Detailed component and widget documentation
- [**Architecture**](docs/architecture.md) - How Terminus works under the hood

## 🎮 Examples

Explore our example applications:

| Example | Description | Run Command |
|---------|-------------|-------------|
| **Hello World** | Simple starter app | `go run ./examples/hello/` |
| **Todo List** | Task management with persistence | `go run ./examples/todo/` |
| **Chat** | Real-time messaging | `go run ./examples/chat/` |
| **Dashboard** | Complex layouts | `go run ./examples/dashboard/` |
| **Widgets** | All widgets showcase | `go run ./examples/widgets/` |
| **Text Input** | Forms with validation | `go run ./examples/textinput/` |
| **Commands** | Advanced command usage | `go run ./examples/commands/` |
| **Layout** | Layout system demo | `go run ./examples/layout/` |
| **Gemini Chat** | AI chat with Google Gemini | `go run ./examples/gemini_chat/` |

All examples run on `http://localhost:8890` by default.

## 🏗️ Architecture

Terminus uses a unique server-side architecture:

```
┌─────────────┐         WebSocket          ┌─────────────┐
│   Browser   │ ◄──────────────────────► │  Go Server  │
│             │         JSON Messages      │             │
│ ┌─────────┐ │                           │ ┌─────────┐ │
│ │  Thin   │ │                           │ │   MVU   │ │
│ │ Client  │ │                           │ │ Engine  │ │
│ └─────────┘ │                           │ └─────────┘ │
└─────────────┘                           └─────────────┘
```

Key benefits:
- **Zero client-side state** - All logic stays in Go
- **Automatic UI updates** - Just update your model
- **Secure by default** - No client-side code execution
- **Easy testing** - Pure functions throughout

## 🛠️ Development Status

### ✅ Completed
- Core MVU engine with component system
- WebSocket communication layer
- Session management
- Full ANSI styling system (16, 256, and RGB colors)
- Efficient line-based diff algorithm
- Complete widget library (TextInput, List, Table, Spinner)
- Layout system with box drawing
- HTTP command helpers
- Comprehensive documentation
- Enhanced JavaScript client with full color support

### 🚧 In Progress
- Performance optimizations
- Mouse support
- Browser compatibility testing
- Security hardening

### 📋 Planned
- Additional widgets (Progress, Select, Tree)
- Hot reload for development
- DevTools browser extension
- Plugin system

## 🤝 Contributing

We welcome contributions! Areas where you can help:

- 📝 Documentation improvements
- 🐛 Bug reports and fixes
- ✨ New widget implementations
- 🎨 Example applications
- 🌍 Internationalization
- ⚡ Performance optimizations

Please see our [Contributing Guide](CONTRIBUTING.md) (coming soon).

## 📄 License

This project is licensed under the Apache 2.0 License.

## 🙏 Acknowledgments

Terminus is inspired by:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework for Go
- [Elm](https://elm-lang.org/) - The MVU architecture
- Terminal emulators and their rich history

---
Happy coding! 🚀
