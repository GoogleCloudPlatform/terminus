# Terminus

A Go framework for building terminal-style user interfaces that run in web browsers. Terminus brings the simplicity and power of terminal UIs to the web, using a Model-View-Update (MVU) architecture similar to Elm.

**NOTE**
This project is a demonstration of vibe coding intended to provide a trustworthy and verifiable example that developers and researchers can use. It is not intended
for use in a production environment.

This is not an officially supported Google product. This project is not
eligible for the [Google Open Source Software Vulnerability Rewards
Program](https://bughunters.google.com/open-source-security).

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

Terminus provides a backend server that streams ANSI escape codes to compatible clients.
You can run a server and connect to it using either a web browser or a native CLI client.

### 1. Run the Terminus Server (e.g., Hello World example)

First, navigate to one of the example applications (e.g., `examples/hello/`) and run the server:

```bash
cd examples/hello/
go run .
```
The server will typically start on `http://localhost:8080` (or another port specified in the example).

### 2. Connect with a Client

#### Option A: Web Browser

Open your web browser and navigate to the address printed by the server (e.g., `http://localhost:8080`). The Terminus application will render directly in your browser.

#### Option B: Native CLI Client

You can also connect using the native CLI client. First, build the client:

```bash
go build -o terminus-cli ../../cmd/terminus/main.go
```

Then, run the client:

```bash
./terminus-cli
```
The CLI client will connect to the running Terminus server and render the application directly in your terminal.

## 📚 Documentation

- [**Getting Started**](docs/getting-started.md) - Quick introduction and your first app
- [**Tutorial**](docs/tutorial.md) - Build a complete todo list application
- [**API Reference**](docs/api.md) - Detailed component and widget documentation
- [**Architecture**](docs/architecture.md) - How Terminus works under the hood

## 🎮 Examples

Explore our example applications:

| Example         | Description                      | Run Command                      |
| --------------- | -------------------------------- | -------------------------------- |
| **Hello World** | Simple starter app               | `go run ./examples/hello/`       |
| **Todo List**   | Task management with persistence | `go run ./examples/todo/`        |
| **Chat**        | Real-time messaging              | `go run ./examples/chat/`        |
| **Dashboard**   | Complex layouts                  | `go run ./examples/dashboard/`   |
| **Widgets**     | All widgets showcase             | `go run ./examples/widgets/`     |
| **Text Input**  | Forms with validation            | `go run ./examples/textinput/`   |
| **Commands**    | Advanced command usage           | `go run ./examples/commands/`    |
| **Layout**      | Layout system demo               | `go run ./examples/layout/`      |
| **Gemini Chat** | AI chat with Google Gemini       | `go run ./examples/gemini_chat/` |

All examples run on `http://localhost:8890` by default.

## 🏗️ Architecture

Terminus now supports both web-based and native CLI frontends, all powered by a single Go backend that streams standard ANSI escape codes.

```
┌─────────────┐         WebSocket          ┌─────────────┐
│   Browser   │ ◄──────────────────────► │  Go Server  │
│ ┌─────────┐ │         Raw ANSI           │             │
│ │ Ghostty │ │        + JSON Control      │ ┌─────────┐ │
│ │   Web   │ │                           │ │   MVU   │ │
│ └─────────┘ │                           │ │ Engine  │ │
│             │                           │ └─────────┘ │
├─────────────┤                           ├─────────────┤
│ Native CLI  │ ◄──────────────────────► │             │
│   Client    │         Raw ANSI           │             │
│ ┌─────────┐ │        + JSON Control      │             │
│ │ x/term  │ │                           │             │
│ └─────────┘ │                           │             │
└─────────────┘                           └─────────────┘
```

Key benefits:

- **Unified Backend**: A single Go server powers both web and native CLI clients.
- **Raw ANSI Streaming**: Efficiently renders terminal UI by sending standard ANSI escape codes.
- **Zero client-side state** - All logic stays in Go
- **Automatic UI updates** - Just update your model
- **Secure by default** - No client-side code execution in the browser, minimal in CLI.
- **Easy testing** - Pure functions throughout

## 🛠️ Development Status

### ✅ Completed

- Core MVU engine with component system
- WebSocket communication layer (raw ANSI streaming)
- Session management
- Full ANSI styling system (16, 256, and RGB colors)
- Efficient line-based diff algorithm
- Complete widget library (TextInput, List, Table, Spinner)
- Layout system with box drawing
- HTTP command helpers
- Comprehensive documentation
- Web frontend with Ghostty-Web terminal (placeholder)
- Native CLI client with raw terminal input/output

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
