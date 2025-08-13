# Terminus API Reference

Terminus is a Go framework for building terminal-style user interfaces that run in web browsers. It uses a Model-View-Update (MVU) architecture similar to Elm.

## Table of Contents

- [Core Components](#core-components)
- [Component Interface](#component-interface)
- [Messages and Commands](#messages-and-commands)
- [Styling](#styling)
- [Widgets](#widgets)
- [Layout](#layout)
- [HTTP Commands](#http-commands)
- [Program](#program)

## Core Components

### Component Interface

Every Terminus application is built using components that implement the `Component` interface:

```go
type Component interface {
    Init() Cmd
    Update(Msg) (Component, Cmd)
    View() string
}
```

#### Methods

##### Init() Cmd
Called when the component is initialized. Returns an optional command to execute.

```go
func (m *MyComponent) Init() terminus.Cmd {
    // Return nil if no initial command needed
    return nil
    
    // Or return a command
    return terminus.Tick(time.Second, func(time.Time) terminus.Msg {
        return TickMsg{}
    })
}
```

##### Update(Msg) (Component, Cmd)
Processes messages and updates the component state. Returns the updated component and an optional command.

```go
func (m *MyComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    switch msg := msg.(type) {
    case terminus.KeyMsg:
        switch msg.Type {
        case terminus.KeyEnter:
            // Handle enter key
            return m, nil
        case terminus.KeyEscape:
            return m, terminus.Quit
        }
    }
    return m, nil
}
```

##### View() string
Renders the component's current state as a string.

```go
func (m *MyComponent) View() string {
    return "Hello, World!"
}
```

## Messages and Commands

### Messages

Messages are events that trigger updates in your components.

#### Built-in Messages

##### KeyMsg
Represents keyboard input:

```go
type KeyMsg struct {
    Type  KeyType
    Runes []rune // For character input
}
```

Key types include:
- `KeyEnter`, `KeySpace`, `KeyBackspace`, `KeyDelete`
- `KeyTab`, `KeyShiftTab`, `KeyEscape`
- `KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`
- `KeyHome`, `KeyEnd`, `KeyPageUp`, `KeyPageDown`
- `KeyCtrlA` through `KeyCtrlZ`
- `KeyF1` through `KeyF12`
- `KeyRunes` (for regular character input)

##### QuitMsg
Signals that the application should quit:

```go
type QuitMsg struct{}
```

##### WindowSizeMsg
Sent when the terminal window is resized:

```go
type WindowSizeMsg struct {
    Width  int
    Height int
}
```

### Commands

Commands are functions that perform side effects and return messages.

```go
type Cmd func() Msg
```

#### Built-in Commands

##### Quit
Returns a command that quits the application:

```go
return terminus.Quit
```

##### Tick
Creates a timer that sends messages at regular intervals:

```go
func Tick(duration time.Duration, fn func(time.Time) Msg) Cmd
```

Example:
```go
return terminus.Tick(time.Second, func(t time.Time) terminus.Msg {
    return TickMsg{Time: t}
})
```

##### Batch
Combines multiple commands into one:

```go
func Batch(cmds ...Cmd) Cmd
```

Example:
```go
return terminus.Batch(
    someCommand(),
    anotherCommand(),
)
```

##### Sequence
Executes commands in order:

```go
func Sequence(cmds ...Cmd) Cmd
```

## Styling

### Style Package

The style package provides a fluent API for text styling:

```go
import "github.com/yourusername/terminusgo/pkg/terminus/style"
```

#### Creating Styles

```go
// Basic styling
styled := style.New().
    Foreground(style.Red).
    Background(style.Blue).
    Bold(true).
    Underline(true).
    Render("Hello, World!")

// Using hex colors
styled := style.New().
    Foreground(style.HexColor("#FF5733")).
    Render("Custom color")

// Using RGB colors
styled := style.New().
    Foreground(style.RGBColor(255, 87, 51)).
    Render("RGB color")
```

#### Available Methods

- `Foreground(Color)` - Set text color
- `Background(Color)` - Set background color
- `Bold(bool)` - Enable/disable bold
- `Italic(bool)` - Enable/disable italic
- `Underline(bool)` - Enable/disable underline
- `Strikethrough(bool)` - Enable/disable strikethrough
- `Blink(bool)` - Enable/disable blinking
- `Reverse(bool)` - Reverse foreground/background
- `Faint(bool)` - Make text faint
- `Render(string)` - Apply style to text

#### Predefined Colors

- Basic: `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White`
- Bright: `BrightBlack`, `BrightRed`, `BrightGreen`, etc.
- Custom: `HexColor(string)`, `RGBColor(r, g, b uint8)`, `ANSI256Color(uint8)`

## Widgets

### TextInput

A single-line text input widget:

```go
import "github.com/yourusername/terminusgo/pkg/terminus/widget"

// Create a text input
input := widget.NewTextInput().
    SetPlaceholder("Enter your name...").
    SetWidth(40).
    OnSubmit(func(value string) terminus.Msg {
        return SubmitMsg{Name: value}
    })

// In your Update method
case terminus.KeyMsg:
    var cmd terminus.Cmd
    input, cmd = input.Update(msg)
    return m, cmd

// In your View method
return input.View()
```

#### Methods

- `SetValue(string)` - Set the current value
- `SetPlaceholder(string)` - Set placeholder text
- `SetWidth(int)` - Set display width
- `SetPrompt(string)` - Set prompt character (default: "> ")
- `SetCursorStyle(style.Style)` - Style the cursor
- `SetTextStyle(style.Style)` - Style the input text
- `SetPlaceholderStyle(style.Style)` - Style placeholder
- `SetPromptStyle(style.Style)` - Style prompt
- `SetValidation(func(string) error)` - Add validation
- `OnChange(func(string) terminus.Msg)` - Handle changes
- `OnSubmit(func(string) terminus.Msg)` - Handle submit
- `Focus()` / `Blur()` - Control focus

### List

A scrollable list widget:

```go
// Create a list
items := []widget.ListItem{
    widget.NewSimpleListItem("1", "First Item"),
    widget.NewSimpleListItem("2", "Second Item"),
    widget.NewSimpleListItem("3", "Third Item"),
}

list := widget.NewList(items).
    SetHeight(10).
    OnSelect(func(item widget.ListItem) terminus.Msg {
        return SelectMsg{ID: item.String()}
    })
```

#### Methods

- `SetItems([]ListItem)` - Set list items
- `SetHeight(int)` - Set visible height
- `SetSelectedIndex(int)` - Set selected item
- `SetItemStyle(style.Style)` - Style normal items
- `SetSelectedStyle(style.Style)` - Style selected item
- `EnableFiltering()` / `DisableFiltering()` - Toggle filtering
- `SetFilter(string)` - Set filter string
- `OnSelect(func(ListItem) terminus.Msg)` - Handle selection

### Table

A data table widget:

```go
// Create a table
headers := []string{"Name", "Age", "City"}
rows := [][]widget.TableCell{
    {
        widget.NewTableCell("Alice"),
        widget.NewTableCell(30),
        widget.NewTableCell("New York"),
    },
    {
        widget.NewTableCell("Bob"),
        widget.NewTableCell(25),
        widget.NewTableCell("London"),
    },
}

table := widget.NewTable(headers, rows).
    SetHeight(10).
    SetColumnWidths([]int{20, 10, 20}).
    EnableSorting()
```

#### Methods

- `SetHeaders([]string)` - Set column headers
- `SetRows([][]TableCell)` - Set table data
- `SetHeight(int)` - Set visible height
- `SetColumnWidths([]int)` - Set column widths
- `EnableSorting()` / `DisableSorting()` - Toggle sorting
- `SetHeaderStyle(style.Style)` - Style headers
- `SetRowStyle(style.Style)` - Style rows
- `SetSelectedStyle(style.Style)` - Style selection
- `SetBorderStyle(style.Style)` - Style borders

### Spinner

An animated loading spinner:

```go
// Create a spinner
spinner := widget.NewSpinner().
    SetStyle(widget.SpinnerDots).
    SetSpeed(100 * time.Millisecond)

// Start animation
cmd := spinner.Tick()
```

#### Spinner Styles

- `SpinnerDots` - Braille dots
- `SpinnerLine` - Rotating line
- `SpinnerCircle` - Circle animation
- `SpinnerSquare` - Square animation
- `SpinnerTriangle` - Triangle animation
- `SpinnerArrow` - Arrow animation
- `SpinnerBouncingBar` - Bouncing bar
- `SpinnerBouncingBall` - Bouncing ball

## Layout

### Box Drawing

```go
import "github.com/yourusername/terminusgo/pkg/terminus/layout"

// Draw a simple box
box := layout.Box{
    Width:  40,
    Height: 10,
    Style:  layout.NormalBorder(),
}
content := box.Render("Hello, World!")

// Double border
box.Style = layout.DoubleBorder()

// Rounded corners
box.Style = layout.RoundedBorder()

// Custom border style
box.Style = layout.BorderStyle{
    Top:    "═",
    Bottom: "═",
    Left:   "║",
    Right:  "║",
    // ... corners
}
```

### Layout Helpers

```go
// Horizontal layout
row := layout.Row(
    layout.Column("Left", 20),
    layout.Column("Center", 30),
    layout.Column("Right", 20),
)

// Vertical layout with padding
content := layout.Pad(2, 1, "Content")

// Center text
centered := layout.Center(80, 24, "Centered Text")
```

## HTTP Commands

### Making HTTP Requests

```go
// Simple GET request
cmd := terminus.Get("https://api.example.com/data")

// POST with JSON
cmd := terminus.Post("https://api.example.com/users", map[string]string{
    "name": "John Doe",
    "email": "john@example.com",
})

// Custom request with headers
cmd := terminus.HTTPRequestWithHeaders(
    terminus.POST,
    "https://api.example.com/auth",
    bytes.NewReader([]byte(`{"username":"admin"}`)),
    map[string]string{
        "Authorization": "Bearer token123",
    },
)
```

### Handling Responses

```go
func (m *MyComponent) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
    switch msg := msg.(type) {
    case terminus.HTTPRequestMsg:
        if msg.Error != nil {
            m.error = msg.Error.Error()
            return m, nil
        }
        
        // Check status
        if msg.IsHTTPError() {
            m.error = fmt.Sprintf("HTTP %d", msg.StatusCode())
            return m, nil
        }
        
        // Parse JSON response
        var data ResponseData
        if err := msg.JSONBody(&data); err != nil {
            m.error = err.Error()
            return m, nil
        }
        
        m.data = data
        return m, nil
    }
    return m, nil
}
```

## Program

### Creating and Running a Program

```go
package main

import (
    "embed"
    "log"
    "github.com/yourusername/terminusgo/pkg/terminus"
)

//go:embed all:static/*
var staticFiles embed.FS

func main() {
    // Create program with options
    program := terminus.NewProgram(
        func() terminus.Component {
            return NewMyComponent()
        },
        terminus.WithAddress(":8080"),
        terminus.WithStaticFiles(staticFiles, "static"),
    )
    
    // Start the program
    if err := program.Start(); err != nil {
        log.Fatal(err)
    }
    
    // Wait for shutdown
    program.Wait()
}
```

### Program Options

- `WithAddress(string)` - Set server address (default: ":8080")
- `WithStaticFiles(embed.FS, string)` - Serve static files

### Static Files

Create a `static` directory with:
- `index.html` - Your HTML template
- `terminus-client.js` - The Terminus client library
- `terminus.css` - Terminal styling

The HTML should include:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
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