# Terminus User Guide

## Introduction
Terminus is a Go framework for building rich, interactive terminal-style user interfaces that run directly in a web browser. It combines the simplicity of text-based interfaces with the accessibility of the web, allowing developers to build complex UIs entirely in Go.

**Key Philosophy:**
*   **Go-Centric:** Logic resides entirely on the server.
*   **Text-First:** UI is rendered as styled text, reminiscent of classic terminal applications.
*   **Component-Based:** Inspired by React and The Elm Architecture (Model-View-Update).
*   **Zero-Config Client:** A minimal, generic JavaScript client handles rendering, requiring no frontend development from the user.

## Core Concepts

### The Elm Architecture Pattern
Terminus follows the Model-View-Update (MVU) pattern:

1.  **Model:** A Go struct representing the state of a component.
2.  **View:** A function that renders the Model into a string (the UI).
3.  **Update:** A function that handles Messages (events) and updates the Model.

### Components
The fundamental building block in Terminus is the `Component` interface. Every part of the UI is a component that manages its own state and rendering.

### Messages and Commands
*   **Messages (`Msg`):** Events that trigger state changes (e.g., key presses, timer ticks, data loaded).
*   **Commands (`Cmd`):** Asynchronous operations (e.g., HTTP requests) that produce Messages upon completion.

## Getting Started

### Installation
```bash
go get github.com/yourusername/terminusgo
```

### Your First Application
A minimal Terminus application requires a root component and a main entry point.

**1. Define the Component**
```go
type HelloModel struct {
    message string
}

func (m HelloModel) Init() terminus.Cmd {
    return nil
}

func (m HelloModel) Update(msg terminus.Msg) (terminus.Model, terminus.Cmd) {
    if keyMsg, ok := msg.(terminus.KeyMsg); ok && keyMsg.String() == "q" {
        return m, terminus.Quit
    }
    return m, nil
}

func (m HelloModel) View() string {
    return "Hello, Terminus! Press 'q' to quit."
}
```

**2. Start the Server**
```go
func main() {
    program := terminus.NewProgram(HelloModel{message: "Welcome"})
    log.Fatal(program.Start(":8080"))
}
```

## Architecture

1.  **Server:** Your Go application runs a WebSocket server.
2.  **Client:** Users visit the URL (e.g., `localhost:8080`) to load a generic HTML/JS client.
3.  **Communication:**
    *   **Server -> Client:** Sends rendered text (the View) to display.
    *   **Client -> Server:** Sends user input (keystrokes) as Messages.
