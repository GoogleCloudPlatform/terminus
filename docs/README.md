# Terminus Documentation

Welcome to the Terminus documentation! Terminus is a Go framework for building terminal-style user interfaces that run in web browsers.

## 📚 Documentation Index

### Getting Started
- [**Getting Started Guide**](getting-started.md) - Quick introduction and your first Terminus app
- [**Tutorial**](tutorial.md) - Step-by-step guide to building a todo list application
- [**Examples**](../examples/README.md) - Sample applications demonstrating various features

### Reference
- [**API Reference**](api.md) - Complete API documentation for all packages
- [**Architecture**](architecture.md) - Technical design and implementation details
- [**User Guide**](../userguide.md) - Comprehensive guide to using Terminus

### Advanced Topics
- [**Widgets**](api.md#widgets) - Pre-built UI components
- [**Styling**](api.md#styling) - Text formatting and colors
- [**Layout**](api.md#layout) - Positioning and box drawing
- [**HTTP Commands**](api.md#http-commands) - Making HTTP requests

## 🚀 Quick Links

- [GitHub Repository](https://github.com/yourusername/terminusgo)
- [Issue Tracker](https://github.com/yourusername/terminusgo/issues)
- [Discussions](https://github.com/yourusername/terminusgo/discussions)
- [Releases](https://github.com/yourusername/terminusgo/releases)

## 📖 Documentation Overview

### For New Users

1. Start with the [Getting Started Guide](getting-started.md) to understand the basics
2. Follow the [Tutorial](tutorial.md) to build your first real application
3. Explore the [Examples](../examples) to see what's possible

### For Developers

1. Read the [API Reference](api.md) for detailed component documentation
2. Understand the [Architecture](architecture.md) for deeper insights
3. Check the [User Guide](../userguide.md) for comprehensive coverage

### For Contributors

1. Review the [Architecture](architecture.md) document
2. Check the [Task List](../tasksterminus.md) for development status
3. Read contribution guidelines (coming soon)

## 🎯 Key Concepts

### Model-View-Update (MVU)

Terminus uses the MVU pattern for predictable state management:

```go
type Component interface {
    Init() Cmd                    // Initialize
    Update(Msg) (Component, Cmd)  // Handle messages  
    View() string                 // Render view
}
```

### Server-Side Rendering

All rendering happens on the server:
- Components render to ANSI-styled text
- Diff algorithm calculates minimal updates
- Updates sent to client via WebSocket
- Client applies changes to DOM

### Component Composition

Build complex UIs from simple components:
- Reusable widgets (TextInput, List, Table, etc.)
- Layout helpers for positioning
- Style system for rich text formatting

## 💡 Example

Here's a simple "Hello, World!" application:

```go
package main

import (
    "github.com/yourusername/terminusgo/pkg/terminus"
    "github.com/yourusername/terminusgo/pkg/terminus/style"
)

type HelloComponent struct{}

func (h *HelloComponent) Init() terminus.Cmd {
    return nil
}

func (h *HelloComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    if _, ok := msg.(terminus.KeyMsg); ok {
        return h, terminus.Quit
    }
    return h, nil
}

func (h *HelloComponent) View() string {
    return style.New().
        Bold(true).
        Foreground(style.Cyan).
        Render("Hello, World! Press any key to exit.")
}

func main() {
    program := terminus.NewProgram(func() terminus.Component {
        return &HelloComponent{}
    })
    
    if err := program.Start(); err != nil {
        panic(err)
    }
    
    program.Wait()
}
```

## 📦 What's Included

### Core Framework
- Component system with MVU pattern
- WebSocket communication layer
- Session management
- Rendering engine with diff algorithm

### Widget Library
- **TextInput** - Single-line text input with validation
- **List** - Scrollable lists with filtering
- **Table** - Data tables with sorting
- **Spinner** - Animated loading indicators

### Utilities
- **Style** - Fluent API for text styling
- **Layout** - Box drawing and positioning helpers
- **HTTP** - Commands for making HTTP requests

### Examples
- **Hello World** - Simplest possible app
- **Todo List** - Task management with persistence
- **Chat** - Real-time messaging
- **Dashboard** - Complex layouts
- And more!

## 🛠️ Development Status

Terminus is actively developed. Check the [Task List](../tasksterminus.md) for current status:

- ✅ Core Framework (Complete)
- ✅ Widget Library (Complete)
- ✅ Examples (Complete)
- ⚠️ Client Implementation (Enhanced)
- ⬜ Production Features (In Progress)

## 🤝 Contributing

We welcome contributions! Areas where you can help:

- 📝 Documentation improvements
- 🐛 Bug reports and fixes
- ✨ New widget implementations
- 🎨 Example applications
- 🌍 Internationalization

## 📄 License

Terminus is open source software licensed under the MIT License.

---

Happy building with Terminus! If you have questions or need help, please:
- Open an [issue](https://github.com/yourusername/terminusgo/issues)
- Start a [discussion](https://github.com/yourusername/terminusgo/discussions)
- Read the [documentation](https://github.com/yourusername/terminusgo/docs)