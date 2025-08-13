# Getting Started with Terminus

Welcome to Terminus! This guide will help you get up and running with your first Terminus application in just a few minutes.

## What is Terminus?

Terminus is a Go framework for building terminal-style user interfaces that run in web browsers. It combines the simplicity of terminal interfaces with the accessibility of the web.

### Key Features

- 🎯 **Simple** - Build UIs with just Go, no JavaScript required
- 🚀 **Fast** - Server-side rendering with efficient updates
- 🎨 **Beautiful** - Rich text styling with ANSI escape sequences
- 📦 **Batteries Included** - Built-in widgets and layouts
- 🔧 **Extensible** - Easy to create custom components

## Installation

### Requirements

- Go 1.16 or later
- A modern web browser (Chrome, Firefox, Safari, Edge)

### Install Terminus

```bash
go get github.com/yourusername/terminusgo
```

## Quick Start

Let's build a simple "Hello, World!" application.

### 1. Create a new project

```bash
mkdir hello-terminus
cd hello-terminus
go mod init hello-terminus
```

### 2. Create the main file

Create `main.go`:

```go
package main

import (
    "github.com/yourusername/terminusgo/pkg/terminus"
    "github.com/yourusername/terminusgo/pkg/terminus/style"
)

// HelloComponent is our main component
type HelloComponent struct {
    message string
}

// Init initializes the component
func (h *HelloComponent) Init() terminus.Cmd {
    return nil
}

// Update handles messages
func (h *HelloComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    switch msg := msg.(type) {
    case terminus.KeyMsg:
        if msg.Type == terminus.KeyEscape {
            return h, terminus.Quit
        }
    }
    return h, nil
}

// View renders the component
func (h *HelloComponent) View() string {
    title := style.New().
        Bold(true).
        Foreground(style.Cyan).
        Render("Welcome to Terminus!")
    
    help := style.New().
        Faint(true).
        Render("Press ESC to quit")
    
    return title + "\n\n" + h.message + "\n\n" + help
}

func main() {
    // Create and run the program
    program := terminus.NewProgram(
        func() terminus.Component {
            return &HelloComponent{
                message: "Hello, World! 👋",
            }
        },
    )
    
    if err := program.Start(); err != nil {
        panic(err)
    }
    
    program.Wait()
}
```

### 3. Run your application

```bash
go run main.go
```

Open your browser to `http://localhost:8080` and you'll see your terminal application!

## Understanding the Basics

### Components

Components are the building blocks of Terminus applications. They implement three methods:

1. **Init()** - Called once when the component starts
2. **Update()** - Handles messages and updates state
3. **View()** - Renders the current state

### Messages

Messages are events that trigger updates:

```go
// Built-in messages
type KeyMsg struct {
    Type  KeyType
    Runes []rune
}

type WindowSizeMsg struct {
    Width  int
    Height int
}

// Custom messages
type MyCustomMsg struct {
    Data string
}
```

### Commands

Commands are functions that perform side effects:

```go
// No command
return h, nil

// Quit command
return h, terminus.Quit

// Timer command
return h, terminus.Tick(time.Second, func(t time.Time) terminus.Msg {
    return TickMsg{Time: t}
})
```

## Adding Interactivity

Let's make our app interactive by adding a counter:

```go
type CounterComponent struct {
    count int
}

func (c *CounterComponent) Init() terminus.Cmd {
    return nil
}

func (c *CounterComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    switch msg := msg.(type) {
    case terminus.KeyMsg:
        switch msg.Type {
        case terminus.KeyUp:
            c.count++
        case terminus.KeyDown:
            c.count--
        case terminus.KeySpace:
            c.count = 0
        case terminus.KeyEscape:
            return c, terminus.Quit
        }
    }
    return c, nil
}

func (c *CounterComponent) View() string {
    return fmt.Sprintf(`
Counter: %s

%s to increase
%s to decrease  
%s to reset
%s to quit`,
        style.New().Bold(true).Render(fmt.Sprintf("%d", c.count)),
        style.New().Foreground(style.Green).Render("↑"),
        style.New().Foreground(style.Red).Render("↓"),
        style.New().Foreground(style.Yellow).Render("SPACE"),
        style.New().Foreground(style.Gray).Render("ESC"),
    )
}
```

## Using Widgets

Terminus includes pre-built widgets for common UI elements:

### Text Input

```go
import "github.com/yourusername/terminusgo/pkg/terminus/widget"

type FormComponent struct {
    nameInput *widget.TextInput
}

func NewFormComponent() *FormComponent {
    return &FormComponent{
        nameInput: widget.NewTextInput().
            SetPlaceholder("Enter your name...").
            SetWidth(40),
    }
}

func (f *FormComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    var cmd terminus.Cmd
    f.nameInput, cmd = f.nameInput.Update(msg)
    return f, cmd
}

func (f *FormComponent) View() string {
    return "Name: " + f.nameInput.View()
}
```

### List

```go
type MenuComponent struct {
    list *widget.List
}

func NewMenuComponent() *MenuComponent {
    items := []widget.ListItem{
        widget.NewSimpleListItem("1", "Start Game"),
        widget.NewSimpleListItem("2", "Options"),
        widget.NewSimpleListItem("3", "Quit"),
    }
    
    return &MenuComponent{
        list: widget.NewList(items).
            SetHeight(5).
            OnSelect(func(item widget.ListItem) terminus.Msg {
                return SelectMsg{Option: item.String()}
            }),
    }
}
```

## Styling

Make your application beautiful with the style package:

### Colors

```go
// Named colors
style.New().Foreground(style.Red).Render("Error!")
style.New().Background(style.Blue).Render("Info")

// RGB colors
style.New().Foreground(style.RGBColor(255, 128, 0)).Render("Orange")

// Hex colors
style.New().Foreground(style.HexColor("#FF5733")).Render("Coral")
```

### Text Decorations

```go
style.New().Bold(true).Render("Bold")
style.New().Italic(true).Render("Italic")
style.New().Underline(true).Render("Underlined")
style.New().Strikethrough(true).Render("Strikethrough")
style.New().Blink(true).Render("Blinking")
```

### Combining Styles

```go
style.New().
    Bold(true).
    Foreground(style.Red).
    Background(style.Yellow).
    Underline(true).
    Render("Important!")
```

## Custom Static Files

To use custom HTML, CSS, or include additional JavaScript:

### 1. Create a static directory

```bash
mkdir static
```

### 2. Add your files

`static/index.html`:
```html
<!DOCTYPE html>
<html>
<head>
    <title>My Terminus App</title>
    <link rel="stylesheet" href="/terminus.css">
</head>
<body>
    <div id="terminal-container">
        <div id="terminal" class="terminal" tabindex="0"></div>
    </div>
    <script src="/terminus-client.js"></script>
</body>
</html>
```

### 3. Embed and serve the files

```go
import "embed"

//go:embed all:static/*
var staticFiles embed.FS

func main() {
    program := terminus.NewProgram(
        func() terminus.Component {
            return NewMyComponent()
        },
        terminus.WithStaticFiles(staticFiles, "static"),
    )
    // ...
}
```

## Next Steps

Now that you've built your first Terminus application:

1. **Explore Examples** - Check out the [examples directory](../examples) for more complex applications
2. **Read the Tutorial** - Follow the [step-by-step tutorial](tutorial.md) to build a todo app
3. **API Reference** - Dive into the [API documentation](api.md) for detailed information
4. **Architecture** - Learn about [how Terminus works](architecture.md) under the hood

## Common Patterns

### Loading Data

```go
func (c *Component) Init() terminus.Cmd {
    return terminus.Batch(
        loadUserData(),
        loadSettings(),
    )
}

func loadUserData() terminus.Cmd {
    return func() terminus.Msg {
        // Fetch data
        user, err := fetchUser()
        return UserLoadedMsg{User: user, Error: err}
    }
}
```

### Periodic Updates

```go
func (c *Component) Init() terminus.Cmd {
    return terminus.Tick(time.Second, func(t time.Time) terminus.Msg {
        return TickMsg{Time: t}
    })
}
```

### HTTP Requests

```go
func (c *Component) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    switch msg := msg.(type) {
    case FetchDataMsg:
        return c, terminus.Get("https://api.example.com/data")
    case terminus.HTTPRequestMsg:
        if msg.Error != nil {
            c.error = msg.Error.Error()
        } else {
            c.data = msg.String()
        }
    }
    return c, nil
}
```

## Tips and Tricks

### 1. Keep Components Small

Break large components into smaller, focused ones for better maintainability.

### 2. Use Type-Safe Messages

Define specific message types instead of using generic interfaces:

```go
// Good
type UserSelectedMsg struct { UserID string }

// Avoid
type GenericMsg struct { Type string; Data interface{} }
```

### 3. Handle Errors Gracefully

Always show meaningful error messages to users:

```go
if err != nil {
    c.error = style.New().
        Foreground(style.Red).
        Render("Error: " + err.Error())
}
```

### 4. Provide Keyboard Shortcuts

Make your app keyboard-friendly:

```go
help := `
Shortcuts:
  j/k - Navigate
  Enter - Select
  / - Search
  ? - Help
  q - Quit
`
```

### 5. Test Your Components

Components are easy to test:

```go
func TestCounter(t *testing.T) {
    comp := &CounterComponent{count: 0}
    
    // Simulate key press
    comp, _ = comp.Update(terminus.KeyMsg{Type: terminus.KeyUp})
    
    // Verify state
    if comp.(*CounterComponent).count != 1 {
        t.Error("Expected count to be 1")
    }
}
```

## Getting Help

- 📖 Read the [documentation](https://github.com/yourusername/terminusgo/docs)
- 💬 Join our [Discord community](#)
- 🐛 Report issues on [GitHub](https://github.com/yourusername/terminusgo/issues)
- 📧 Email support@terminus.dev

Happy building with Terminus! 🚀